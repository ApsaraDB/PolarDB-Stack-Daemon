/* 
*Copyright (c) 2019-2021, Alibaba Group Holding Limited;
*Licensed under the Apache License, Version 2.0 (the "License");
*you may not use this file except in compliance with the License.
*You may obtain a copy of the License at

*   http://www.apache.org/licenses/LICENSE-2.0

*Unless required by applicable law or agreed to in writing, software
*distributed under the License is distributed on an "AS IS" BASIS,
*WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*See the License for the specific language governing permissions and
*limitations under the License.
 */


package timer

import (
	"fmt"
	config "github.com/ApsaraDB/PolarDB-Stack-Daemon/cmd/daemon/app/config"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	utils "github.com/ApsaraDB/PolarDB-Stack-Daemon/polar-controller-manager"
	wwidutil "github.com/ApsaraDB/PolarDB-Stack-Daemon/polar-controller-manager/util"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

const PrintPeriodMinute = 1 * time.Minute

type printPort struct {
	client    *clientset.Clientset
	configMap string
}

type rangePort struct {
	Start int
	End   int
}

func StartPrintPort(client *clientset.Clientset, stop <-chan struct{}) {
	defer utilruntime.HandleCrash()
	klog.Infof("Starting port usage controller, MpdControllerConfigMapName:%s", config.Conf.MpdControllerConfigMapName)
	defer klog.Infof("Shutting port usage controller")
	host, _ := os.Hostname()

	w := &printPort{client, fmt.Sprintf("cloud-provider-port-usage-%s", host)}

	wait.Until(w.PrintPort, PrintPeriodMinute, stop)
}

func (p *printPort) checkSwitchStatus() bool {
	controllerConf, err := utils.GetControllerConfig()
	if err != nil {
		klog.Errorf("StartPrintPort get controller-config error:[%v]", err)
		return false
	}
	return controllerConf.EnablePrintPort
}

func (p *printPort) markUnPrintPort() []rangePort {
	var res []rangePort

	cm, err := p.client.CoreV1().ConfigMaps("kube-system").Get(config.Conf.MpdControllerConfigMapName, metav1.GetOptions{})
	if err != nil && apierrors.IsNotFound(err) {
		return res
	} else if err != nil {
		klog.Errorf("get %s configmap failed: %s", p.configMap, err.Error())
		return res
	}
	ams := cm.Annotations

	for _, v := range ams {
		rangx := parseRange(v)

		if rangx.Start == 0 {
			continue
		}
		res = append(res, rangx)
	}
	return res
}

func buildAwkShellParts(rangx []rangePort) string {
	if len(rangx) < 1 {
		return ""
	}
	var cmdPart []string
	for _, r := range rangx {
		//(int($1)>=5400 && int($1)<=5800))
		cmdPart = append(cmdPart, fmt.Sprintf("int($1)>=%v && int($1)<=%v", r.Start, r.End))
	}
	return fmt.Sprintf("&& (%v)", strings.Join(cmdPart, "||"))
}

func parseRange(rangx string) rangePort {
	rangx = strings.TrimSpace(rangx)
	x, xE := strconv.Atoi(rangx)
	if xE == nil {
		return rangePort{x, x}
	}
	parts := wwidutil.ReduceArray(strings.Split(rangx, "-"))
	if len(parts) != 2 {
		return rangePort{0, 0}
	}
	l, lE := strconv.Atoi(parts[0])
	if lE != nil {
		return rangePort{0, 0}
	}
	r, rE := strconv.Atoi(parts[1])
	if rE != nil {
		return rangePort{0, 0}
	}

	if l < 0 {
		l = 0
	}
	if r < 0 {
		r = 0
	}

	if l <= r {
		return rangePort{l, r}
	} else {
		return rangePort{r, l}
	}
}

// PrintPort
/**
* @Title:  已用端口输出到configMag
* @Description:
*
*	除了已占用端口采用最新的端口扫描逻辑
* 	其余延用原有逻辑
* 	去除了 isPrintPort 的验证
 */
func (p *printPort) PrintPort() {
	if !p.checkSwitchStatus() {
		klog.Infof("PrintPort switch is off , skip port print!")
		return
	}

	cm, err := p.client.CoreV1().ConfigMaps("kube-system").Get(p.configMap, metav1.GetOptions{})
	if err != nil && apierrors.IsNotFound(err) {
		cm = &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      p.configMap,
				Namespace: "kube-system",
			},
		}
		cm, err = p.client.CoreV1().ConfigMaps("kube-system").Create(cm)
		if err != nil {
			klog.Errorf("create %s configmap failed: %s", p.configMap, err.Error())
			return
		}
	} else if err != nil {
		klog.Errorf("get %s configmap failed: %s", p.configMap, err.Error())
		return
	}

	rangePort := p.markUnPrintPort()
	klog.Infof("Print Port range: %v", rangePort)

	var newData = make(map[string]string)

	// 采用新的端口扫描判断是否被点用，除已占用端口采用最的逻辑，其余延用原有逻辑，去除了 isPrintPort 的验证
	alreadyUsePorts := scanRangePort(rangePort)

	for _, alreadyUsePort := range alreadyUsePorts {
		portKey := strconv.Itoa(alreadyUsePort)
		if value, ok := newData[portKey]; ok {
			newData[portKey] = value
		} else {
			newData[portKey] = "the port already use"
		}
	}

	var isChanged bool

	// 判断是否有元素需要删除
	for key, value := range cm.Data {
		if _, ok := newData[key]; !ok {
			isChanged = true
			klog.Infof("[timer port usage] - %s: %s", key, value)
			break
		}
	}
	if !isChanged {
		for key, value := range newData {
			if data, ok := cm.Data[key]; !ok || value != data {
				isChanged = true
				klog.Infof("[timer port usage] + %s: %s", key, value)
				break
			}
		}
	}
	if isChanged {
		if cm.Annotations == nil {
			cm.Annotations = make(map[string]string)
		}
		cm.Annotations["updateTimestamp"] = time.Now().UTC().Format("2006-01-02T15:04:05Z")
		cm.Data = newData
		_, err = p.client.CoreV1().ConfigMaps("kube-system").Update(cm)
		if err != nil {
			klog.Errorf("[timer port usage] update configmap failed: %v", err)
			return
		}
	} else {
		klog.Infof("[timer port usage] %s configmap is not changed.", p.configMap)
	}
}

/**
 * @Title: 扫描区域端口是否可用
 * @Description:
 *
 *	指定扫描端口区域数组：rangePorts []rangePort
 *	将返回已被占用的端口list：alreadyUsePort
 */
func scanRangePort(rangePorts []rangePort) (alreadyUsePort []int) {
	defer func() {
		if err := recover(); err != nil {
			klog.Errorf("scanRangePort scan range port error %v", err)
		}
	}()

	klog.Infof("scanRangePort begin scan range %#v", rangePorts)
	for _, rangePort := range rangePorts {
		for port := rangePort.Start; port < rangePort.End; port++ {
			if !isPortAvailable(port) {
				alreadyUsePort = append(alreadyUsePort, port)
			}
		}
	}
	klog.Infof("scanRangePort range port scan result:%#v", alreadyUsePort)
	return
}

/**
 * @Title: 端口是否可用
 * @Description:
 *
 *	采用go socket对象尝试监听所在服务器指定端口
 *	- 若监听成功，说明该端口此时未被其它应用使用
 *	- 若监听失败，说明该端口已被占用
 */
func isPortAvailable(port int) bool {
	address := fmt.Sprintf("%s:%d", "0.0.0.0", port)
	listener, err := net.Listen("tcp", address)

	if err != nil {
		return false
	}

	defer listener.Close()
	return true
}
