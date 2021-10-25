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

package core_version

import (
	"fmt"
	config "github.com/ApsaraDB/PolarDB-Stack-Daemon/cmd/daemon/app/config"
	"github.com/ApsaraDB/PolarDB-Stack-Daemon/polar-controller-manager/util"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"strings"
	"time"
)

// StartCheckCoreVersion
/**
 * @Title:  StartCheckCoreVersion
 * @Description: 开始等待 检查 core version 的指令
 **/
func StartCheckCoreVersion(client *clientset.Clientset) {
	polarStackDaemonLabels = config.Conf.PolarStackDaemonPodLabels
	coreVersionConfigMapLabel = config.Conf.CoreVersionConfigMapLabel

	for {
		klog.Infof("%s StartCheckCoreVersion ", logInfoTarget)

		klog.Infof("%s polarStackDaemonLabels = %s ", logInfoTarget, polarStackDaemonLabels)
		klog.Infof("%s coreVersionConfigMapLabel = %s ", logInfoTarget, coreVersionConfigMapLabel)
		go startCheckCoreVersion(client)

		request := <-CheckRequestQueue
		// 请求仅为检测当前节点，再广播通知其它 polarstack-daemon 节点
		if request == SingleCheckCoreVersionOperatorType {
			continue
		}

		startNotifyOtherNode(client)

		klog.Infof("%s StartCheckCoreVersion current host name:%s requestType:%s", logInfoTarget, hostName, request)
	}
}

// startNotifyOtherNode
/**
 * @Title:  startNotifyOtherNode
 * @Description: 开始通知其它节占主机
 **/
func startNotifyOtherNode(client *clientset.Clientset) {
	defer func() {
		if err := recover(); err != nil {
			klog.Errorf("%s failed to startNotifyOtherNode err:%v", logInfoTarget, err)
		}
	}()
	var httpClient = &util.HttpClient{Timeout: 3 * time.Second}

	// 获取 polarstack-daemon 节点的 ip
	ips := getPolarStackDaemonNodeIps(client)
	if nil == ips {
		klog.Infof("%s StartCheckCoreVersion failed get daemon ips = nil", logInfoTarget)
		return
	}
	for _, ip := range ips {
		klog.Infof("%s hostName:[%s] get daemon ips", logInfoTarget, hostName)
		header := make(map[string]string)
		type source struct {
			HostName string
		}

		url := "/innerCheckCoreVersion"
		httpClient.Host = fmt.Sprintf("https://%s:%d", ip, config.Conf.Port)

		resp, err := httpClient.HttpsPost(url, header, &source{
			HostName: hostName,
		})

		if err != nil {
			klog.Errorf("%s failed to post ip:%s, url:%s, err:%s", logInfoTarget, ip, url, err.Error())
			continue
		}

		klog.Infof("%s success! ip:%s url:%s, httpStatusCode:[%d]", logInfoTarget, ip, url, resp.StatusCode)
	}
}

// startCheckCoreVersion
/**
 * @Title:  startCheckCoreVersion
 * @Description: 开始检查内核版本（具体实现）
 **/
func startCheckCoreVersion(client *clientset.Clientset) {
	defer func() {
		if err := recover(); err != nil {
			klog.Errorf("%s startCheckCoreVersion failed: err:%s", logInfoTarget, err)
		}
	}()

	// 获取所有 core version 的 configMap
	coreVersions, err := getAllCoreVersionConfigMapByLabels(client, coreVersionConfigMapLabel)
	if err != nil {
		klog.Errorf("%s getAllCoreVersionConfigMapByLabels failed. err:%v", logInfoTarget, err)
		return
	}

	klog.Infof("%s success get [%d] core version config, ready to check now.", logInfoTarget, len(coreVersions.Items))

	existingVersions := getExistingVersions(coreVersions)

	klog.Infof("%s check done. now will update the host core version configMap. %s", logInfoTarget, existingVersions)
	versionConfigMap, err := getHostCoreVersionConfigMap(client)
	if err != nil {
		klog.Errorf("%s failed to get core version configMap. err:%s", logInfoTarget, err.Error())
		return
	}
	// 将末尾,号去除
	if len(existingVersions) > 0 {
		existingVersions = strings.TrimRight(existingVersions, ",")
	} else {
		klog.Warningf("%s existingVersions is empty.", logInfoTarget)
	}

	versionConfigMap.Data["existingVersions"] = existingVersions
	versionConfigMap.Data["checkTime"] = time.Now().Format(timeFormat)

	_, err = updateHostCoreVersionConfigMap(client, versionConfigMap)
	if err != nil {
		klog.Errorf("%s failed to update host core version configMap. err:%s", logInfoTarget, err.Error())
		return
	}
	klog.Infof("%s Success! already update [%s].Data to versions:%s", logInfoTarget, versionConfigMap.Name, existingVersions)
}

// getExistingVersions
/**
 * @Title:  getExistingVersions
 * @Description: 获取本主机上存在的 core version
 **/
func getExistingVersions(coreVersions *v1.ConfigMapList) (existingVersions string) {
	util.ClearImagesCache()
	for _, coreVersion := range coreVersions.Items {
		images, err := getAllImagesByVersionConfigMap(&coreVersion)
		if err != nil {
			klog.Errorf("%s failed to getAllImagesByVersionConfigMap err:%s", logInfoTarget, err.Error())
			continue
		}

		if len(images) == 0 {
			klog.Errorf("%s coreVersion.Name:%s configMap.Data images is empty, maybe not a core version configMap.", logInfoTarget, coreVersion.Name)
			continue
		}

		coreVersionIsExists := true
		for _, image := range images {
			if !util.ImageIsExists(image, logInfoTarget) {
				klog.Infof("%s image[%s] does not exist on current host, configMap:%s", logInfoTarget, image, coreVersion.Name)
				coreVersionIsExists = false
				break
			}
		}

		// 所有镜像存在时，汇总后更新到主机的 configMap 中
		if coreVersionIsExists {
			// get core version
			versionName, _ := getCoreVersionNameByVersionConfigMap(&coreVersion)
			if len(versionName) == 0 {
				klog.Errorf("%s why this version name is empty? configMap.Name is: %s", logInfoTarget, coreVersion.Name)
				continue
			}
			existingVersions += versionName + ","
			klog.Infof("%s The version name %s exists on current host", logInfoTarget, versionName)
		} else {
			klog.Infof("%s this version does not exist on current host, configMap.Name:%s", logInfoTarget, coreVersion.Name)
		}
	}
	util.ClearImagesCache()
	return
}

// GetDaemonNodeIps
/**
 * @Title:  GetDaemonNodeIps
 * @Description: 获取 polarstack-daemon 布属的节点 ip 列表
 **/
func getPolarStackDaemonNodeIps(client *clientset.Clientset) (ips []string) {
	polarStacks, err := client.CoreV1().Pods(NameSpace).List(metav1.ListOptions{LabelSelector: polarStackDaemonLabels})
	if nil != err {
		klog.Errorf("%s getPolarStackDaemonNodeIps failed err: %s", logInfoTarget, err)
		return
	}
	for _, daemon := range polarStacks.Items {
		if strings.ToLower(hostName) != strings.ToLower(daemon.Spec.NodeName) {
			ips = append(ips, daemon.Status.PodIP)
		} else {
			klog.Infof("current host：[%s] not need to notify", daemon.Spec.NodeName)
		}
	}
	return
}
