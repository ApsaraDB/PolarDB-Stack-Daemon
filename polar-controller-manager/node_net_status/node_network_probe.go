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

package node_net_status

import (
	"encoding/json"
	"fmt"
	"github.com/ApsaraDB/PolarDB-Stack-Daemon/polar-controller-manager/util"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientSet "k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	alicloud "github.com/ApsaraDB/PolarDB-Stack-Daemon/polar-controller-manager"
	utils "github.com/ApsaraDB/PolarDB-Stack-Daemon/polar-controller-manager"
)

// NodeClientNetworkUnavailable means that client network of the node is not correctly configured.
const NodeClientNetworkUnavailable v1.NodeConditionType = "NodeClientNetworkUnavailable"

// NodeClientIP means status of the node client ip.
const NodeClientIP v1.NodeConditionType = "NodeClientIP"

// NodeOobIP means status of the node oob ip.
const NodeOobIP v1.NodeConditionType = "NodeOobIP"

// NodeRefreshFlag means status of the node.
const NodeRefreshFlag v1.NodeConditionType = "NodeRefreshFlag"

var ProbeRunnerCounter uint64 = 0

type PolarNodeNetworkProbe struct {
	KubeClient *clientSet.Clientset
	NodeName   string
	cnt        uint64

	nodeSshConn *alicloud.SSHConnection

	ClientCardName string

	isInit bool
}

var (
	hybridDeploySetting *HybridDeploySetting
	once                sync.Once
	netProbeJobPeriod   = 3 * time.Second
)

type HybridDeploySetting struct {
	IsCheckOObIP  bool
	DisableSanCmd bool
	Err           error
}

func __ProbeCounterGetter() uint64 {
	return atomic.AddUint64(&ProbeRunnerCounter, 1)
}

func formatLogLevel(cnt uint64, expectLogLevel int) klog.Level {
	//定在每10个周期打印一次
	if cnt%10 == 1 {
		return klog.Level(expectLogLevel)
	}
	return klog.Level(expectLogLevel + 1)
}

func checkJobHandleCrash() {
	if r := recover(); r != nil {
		klog.Error("checkJobHandleCrash!")
		for _, fn := range utilruntime.PanicHandlers {
			fn(r)
		}
	}
}

func StartNodeNetworkProbe(client *clientset.Clientset, stop <-chan struct{}) {
	period := netProbeJobPeriod
	if innerPeriod := os.Getenv("ProbeJobPeriod"); len(innerPeriod) > 0 {
		if periodInt, err := strconv.ParseInt(innerPeriod, 10, 64); err != nil {
			period = time.Second * time.Duration(periodInt)
		}
	}
	klog.V(5).Infof("ProbeJob period: %0.fs", period.Seconds())
	//
	hostName, _ := os.Hostname()
	klog.V(5).Infof("checkJob : hostName: %s", hostName)

	probe := NewPolarNodeNetworkProbe(client, hostName)

	iErr := probe.Init()
	if iErr != nil {
		klog.V(5).Infof("ProbeCheckJob init err:%v", iErr)
	}

	wait.Until(func() {
		// recover crash
		defer checkJobHandleCrash()

		klog.V(6).Infof("ProbeCheckJob period start: %s", time.Now())
		err := probe.DoNetWorkProbe()
		if err != nil {
			klog.V(5).Infof("ProbeCheckJob err:%v", err)
		}
		return
	}, period, stop)
}

func NewPolarNodeNetworkProbe(client *clientSet.Clientset, nodeName string) *PolarNodeNetworkProbe {
	return &PolarNodeNetworkProbe{
		KubeClient: client,
		NodeName:   nodeName,
	}
}

type StorageStatus struct {
	Available bool
	Status    string
}

var polarNodeNetworkProbeLocker = sync.Mutex{}

func (probe *PolarNodeNetworkProbe) Init() error {
	if probe.KubeClient == nil {
		return fmt.Errorf("kubeClient is nil")
	}

	if probe.NodeName == "" {
		return fmt.Errorf("node is nil")
	}

	if probe.isInit == true {
		return nil
	}

	polarNodeNetworkProbeLocker.Lock()
	defer polarNodeNetworkProbeLocker.Unlock()

	err := probe.__initSSH()
	if err != nil {
		klog.Errorf("probe.__initSSH() for node %s err: %v", probe.NodeName, err)
		return err
	}

	probe.isInit = true

	return nil
}

func (probe *PolarNodeNetworkProbe) DoNetWorkProbe() error {
	if probe.isInit != true {
		err := probe.Init()
		if err != nil {
			klog.Errorf("PolarNodeNetworkProbe init err: %v", err)
			return err
		}
	}

	probe.cnt = __ProbeCounterGetter()

	err := probe.enSureSshConn(probe.cnt)
	if err != nil {
		klog.Errorf("%v", err)
		return err
	}
	start := time.Now()

	defer func() {
		total := time.Now().Sub(start).Seconds()
		if total >= 2 {
			klog.V(4).Infof("[%d]DoNetWorkProbe task done ,spend %v s", probe.cnt, total)
		} else {
			klog.V(5).Infof("[%d]DoNetWorkProbe task done ,spend %v s", probe.cnt, total)
		}
	}()

	hybridDeploySetting = probe.GetHybridDeploySetting()
	nodeList, err := probe.KubeClient.CoreV1().Nodes().List(v12.ListOptions{})
	if err != nil {
		klog.Errorf("PolarNodeNetworkProbe get nodeList err: %v", err)
		return err
	}
	for _, node := range nodeList.Items {
		if probe.NodeName == node.Name {
			// local node
			err = probe.updateNodeClientNetworkCondition(&node)
			if err != nil {
				klog.Errorf("updateNodeClientCondition err: %v", err)
			}
			err = probe.updateNodeClientIPCondition(&node)
			if err != nil {
				klog.Errorf("updateNodeClientCondition err: %v", err)
			}

			if hybridDeploySetting.Err != nil || (hybridDeploySetting.Err == nil && hybridDeploySetting.IsCheckOObIP) {
				err = probe.updateNodeOobCondition(&node)
				if err != nil {
					klog.Errorf("updateNodeOobCondition err: %v", err)
				}
			}

			err = probe.updateNodeRefreshFlagCondition(&node)
			if err != nil {
				klog.Errorf("updateNodeRefreshFlagCondition err: %v", err)
			}
		}
	}

	return nil
}

func (probe *PolarNodeNetworkProbe) enSureSshConn(cnt uint64) error {
	if probe.nodeSshConn == nil {
		probe.nodeSshConn = alicloud.NewSSHConnectionByHost(probe.NodeName, fmt.Sprintf("cnt=%d", cnt), "NetProbe")
	} else {
		probe.nodeSshConn.TagStr = "[" + strings.Join([]string{
			fmt.Sprintf("cnt=%d", cnt), "NetProbe",
		}, "|") + "]"
		probe.nodeSshConn.ResetTime()
	}

	if !probe.nodeSshConn.TestAlive() {
		initErr := probe.nodeSshConn.Init()
		if initErr != nil {
			klog.Errorf("%d runSsh conn.Init hostIP: [%s] failed on build connect:[%s]", cnt, probe.NodeName, initErr)
			return initErr
		}
	}
	return nil
}

func (probe *PolarNodeNetworkProbe) updateNodeClientNetworkCondition(node *v1.Node) error {
	if probe.NodeName != node.Name {
		return fmt.Errorf("updateNodeClientNetworkCondition : not local node ")
	}

	isClientOk, reason, msg := probe.GetNodeClientStatus()

	klog.V(6).Infof("updateNodeClientNetworkCondition node: %s, isClientOk-value: %v, msg:%s",
		node.Name, isClientOk, msg)

	netConStatus := v1.ConditionUnknown
	if isClientOk {
		netConStatus = v1.ConditionFalse
	} else {
		netConStatus = v1.ConditionTrue
	}
	newCondition := &v1.NodeCondition{
		Type:    NodeClientNetworkUnavailable,
		Status:  netConStatus,
		Reason:  reason,
		Message: msg,
	}
	if err := probe.createOrUpdateNodeCondition(node, newCondition); err != nil {
		return err
	}

	return nil
}

func (probe *PolarNodeNetworkProbe) updateNodeClientIPCondition(node *v1.Node) error {
	if probe.NodeName != node.Name {
		return fmt.Errorf("updateNodeOobCondition : not local node ")
	}

	clientIPCond := GetNodeCondition(node, NodeClientIP)
	if clientIPCond != nil {
		subTime := time.Now().Sub(clientIPCond.LastHeartbeatTime.Time)
		if subTime.Hours() <= 1 && clientIPCond.Status == v1.ConditionTrue && clientIPCond.Reason == probe.ClientCardName {
			//状态，网卡未变，可以1小时更新一次。
			klog.V(5).Infof("node %s cond %v[status=%v], last update %v [%v/%v], skip this times check!!", node.Name, NodeClientIP, clientIPCond.Status, clientIPCond.LastHeartbeatTime, subTime.Seconds(), 1*60*60)
			return nil
		}
	}

	newCondition := &v1.NodeCondition{Type: NodeClientIP, Reason: probe.ClientCardName}
	ip, err := probe.getNodeClientIp()
	if err != nil {
		klog.Errorf("Failed to get ipv4 of client net card of node %s, err:%v", node.Name, err)
		newCondition.Status = v1.ConditionFalse
		if clientIPCond != nil && clientIPCond.Message != "0.0.0.0" && len(clientIPCond.Message) > 0 {
			newCondition.Message = clientIPCond.Message
		} else {
			newCondition.Message = "0.0.0.0"
		}
	} else {
		newCondition.Status = v1.ConditionTrue
		newCondition.Message = ip
	}
	if err = probe.createOrUpdateNodeCondition(node, newCondition); err != nil {
		return err
	}

	return nil
}

func (probe *PolarNodeNetworkProbe) getNodeClientIp() (ip string, err error) {
	return probe.getIpByNetCardName(probe.ClientCardName)
}

func (probe *PolarNodeNetworkProbe) getIpByNetCardName(netCardName string) (ip string, err error) {
	// add retry times
	var nic *net.Interface
	var addrs []net.Addr
	for i := 0; i < 3; i++ {
		if i > 0 {
			// if retry then sleep 10 second.
			time.Sleep(10 * time.Second)
		}
		nic, err = net.InterfaceByName(netCardName)
		if err != nil {
			klog.Warningf("failed to read netCard:%s, err:%v", netCardName, err)
			continue
		}

		addrs, err = nic.Addrs()
		if err != nil {
			klog.Warningf("failed to get addrs of netCard:%s, err:%v", netCardName, err)
			continue
		}
		length := len(addrs)
		if length == 0 {
			err = fmt.Errorf("[%d] failed to get any ip from nic %s", i+1, netCardName)
			klog.Errorf("addrs of netCard:%s is empty", netCardName)
			continue
		}
		ip, _, _ := net.ParseCIDR(addrs[0].String())
		ip4 := ip.To4()
		if ip4 == nil {
			if length > 1 {
				klog.Warningf("the first ip of netCard %s is %v, it is not ipv4, will use the second one, already try [%d] times",
					netCardName, ip.String(), i+1)
				ip, _, _ = net.ParseCIDR(addrs[1].String())

				return ip.String(), nil
			} else {
				klog.Warningf("the first ip of netCard %s is %v, it is not ipv4, but this netCard only has this one, already try [%d] times",
					netCardName, ip.String(), i+1)

				err = fmt.Errorf("the first ip of netCard %s is %v, it is not ipv4, but this netCard only has this one, already try [%d] times",
					netCardName, ip.String(), i+1)
				continue
			}
		} else {
			klog.Infof("the first ip of netCard %s is %v, it is ipv4, retry index:[%d]",
				netCardName, ip.String(), i+1)
			return ip.String(), nil
		}
	}

	return
}

func (probe *PolarNodeNetworkProbe) updateNodeOobCondition(node *v1.Node) error {
	if probe.NodeName != node.Name {
		return fmt.Errorf("updateNodeOobCondition : not local node ")
	}

	oobCond := GetNodeCondition(node, NodeOobIP)
	if oobCond != nil {
		subTime := time.Now().Sub(oobCond.LastHeartbeatTime.Time)
		if subTime.Hours() <= 1 && oobCond.Status == v1.ConditionTrue {
			klog.V(5).Infof("node %s cond %v[status=%v], last update %v [%v/%v], skip this times check!!", node.Name, NodeOobIP, oobCond.Status, oobCond.LastHeartbeatTime, subTime.Seconds(), 1*60*60)
			return nil
		}
	}

	newCondition := &v1.NodeCondition{Type: NodeOobIP}
	ip, err := probe.getNodeOobIp()
	if err != nil {
		newCondition.Status = v1.ConditionFalse
		newCondition.Reason = fmt.Sprintf("error: %v", err)
		if oobCond != nil && oobCond.Message != "0.0.0.0" && len(oobCond.Message) > 0 {
			//异常情况下，不修改oob ip
			newCondition.Message = oobCond.Message
		} else {
			newCondition.Message = "0.0.0.0"
		}

	} else {
		newCondition.Status = v1.ConditionTrue
		newCondition.Reason = "PowerOn"
		newCondition.Message = ip
	}
	if err = probe.createOrUpdateNodeCondition(node, newCondition); err != nil {
		return err
	}

	return nil
}

func (probe *PolarNodeNetworkProbe) getNodeOobIp() (ip string, err error) {
	cmd := `ipmitool lan print 1 | grep "IP Address" | grep -v Source| awk '{print $4}'`
	output, stderr, err := utils.RunSSHNoPwdCMD(cmd, "127.0.0.1", "check node oob status")
	if err != nil || util.NotSshLoginWarnMsg(stderr) {
		if util.GetShellErrorExitCode(err) == "1" {
			klog.V(5).Infof("[check node oob status] shell return exit code: cmd: [%s], out:[%s] err:[%s](%v)", cmd, ip, stderr, err)
		} else {
			klog.Errorf("[check node oob status] run cmd failed: out:[%s] cmd:[%s] err:[%s](%v)", ip, cmd, stderr, err)
			return ip, err
		}
	}

	ip = strings.TrimSpace(output)
	if len(ip) == 0 {
		err := fmt.Errorf("ip is null")
		klog.Errorf("Failed to get node oob ip: %v", err)
		return "", err
	}

	if ipCheck := net.ParseIP(ip); ipCheck == nil {
		err := fmt.Errorf("ip %s is unavailable", ip)
		klog.Errorf("Failed to get node oob ip: %v", err)
		return "", err
	}

	return ip, nil
}

func (probe *PolarNodeNetworkProbe) updateNodeRefreshFlagCondition(node *v1.Node) error {
	if probe.NodeName != node.Name {
		return fmt.Errorf("updateNodeRefreshFlagCondition : not local node ")
	}

	newCondition := &v1.NodeCondition{Type: NodeRefreshFlag}
	newCondition.Status = v1.ConditionTrue
	newCondition.Reason = "Refresh"
	newCondition.Message = "Refresh"

	if err := probe.createOrUpdateNodeCondition(node, newCondition); err != nil {
		return err
	}

	return nil
}
func (probe *PolarNodeNetworkProbe) createOrUpdateNodeCondition(
	node *v1.Node, condition *v1.NodeCondition) error {

	condition.LastHeartbeatTime = v12.Now()
	oldCondition := GetNodeCondition(node, condition.Type)
	if oldCondition != nil {
		condition.LastTransitionTime = oldCondition.LastTransitionTime
	} else {
		condition.LastTransitionTime = condition.LastHeartbeatTime
	}

	if err := probe.updateNodeCondition(node, condition); err != nil {
		return err
	}
	return nil
}

func (probe *PolarNodeNetworkProbe) updateNodeCondition(
	node *v1.Node, cond *v1.NodeCondition) error {

	uErr := _SetNodeCondition(probe.KubeClient, node.Name, *cond)

	if uErr != nil {
		klog.Errorf("updateNodeCondition node %s, condition %v, err: %v", node.Name, cond.Type, uErr)
		return uErr
	}

	if cond.Status == v1.ConditionTrue {
		//异常情况下，日志升级打印
		klog.V(4).Infof(
			"updateNodeCondition node: %s, Type:%v ,value: %v", node.Name, cond.Type, cond.Status)
	} else {
		klog.V(formatLogLevel(probe.cnt, 5)).Infof(
			"updateNodeCondition node: %s, Type:%v, value: %v", node.Name, cond.Type, cond.Status)
	}

	return nil
}

func _SetNodeCondition(c *clientSet.Clientset, node string, condition v1.NodeCondition) error {
	generatePatch := func(condition v1.NodeCondition) ([]byte, error) {
		raw, err := json.Marshal(&[]v1.NodeCondition{condition})
		if err != nil {
			return nil, err
		}
		return []byte(fmt.Sprintf(`{"status":{"conditions":%s}}`, raw)), nil
	}
	condition.LastHeartbeatTime = v12.NewTime(time.Now())
	condition.LastTransitionTime = v12.NewTime(time.Now())
	patch, err := generatePatch(condition)
	if err != nil {
		return nil
	}
	_, err = c.CoreV1().Nodes().PatchStatus(string(node), patch)
	return err
}

func GetNodeCondition(node *v1.Node, conType v1.NodeConditionType) *v1.NodeCondition {
	for _, c := range node.Status.Conditions {
		if c.Type == conType {
			return &c
		}
	}
	return nil
}

//严禁在Init之外调用
func (probe *PolarNodeNetworkProbe) __initSSH() error {

	cardName := probe.__GetClientNetCardName()

	if cardName == "" {
		return fmt.Errorf("get client card info err: nil")
	}

	probe.ClientCardName = cardName

	node, nErr := probe.KubeClient.CoreV1().Nodes().Get(probe.NodeName, v12.GetOptions{})

	if nErr != nil {
		klog.Errorf("Get the node %s err:%v", probe.NodeName, nErr)
		return nErr
	}

	nodeAddress := _GetNodeAddress(node, "init-"+node.Name)
	probe.nodeSshConn = alicloud.NewSSHConnectionByHost(nodeAddress, "init-"+probe.NodeName+"-Probe")
	cErr := probe.nodeSshConn.Init()
	if cErr != nil {
		klog.Errorf("init node %s ssh connection err: %v", probe.NodeName, cErr)
		return cErr
	}
	return nil
}

func (probe *PolarNodeNetworkProbe) GetNodeClientStatus() (bool, string, string) {
	return probe.nodeClientAvailable("NodeClientCheck")
}

func (probe *PolarNodeNetworkProbe) __GetClientNetCardName() string {
	netConfig, err := probe.KubeClient.CoreV1().ConfigMaps("kube-system").Get("ccm-config", v12.GetOptions{})
	defaultBondName := "bond1"
	if err != nil {
		klog.Errorf("----system try to get config client network card err,so use default card %s ! %v", defaultBondName, err)
		return defaultBondName
	}

	netcardName, ok := netConfig.Data["NET_CARD_NAME"]
	if !ok {
		return defaultBondName
	}

	if netcardName == "" {
		return defaultBondName
	}

	return netcardName
}

func _FormatSanOutput(outString string) *map[string]*StorageStatus {
	org := strings.TrimSpace(outString)
	if org == "" {
		return nil
	}

	lines := strings.Split(org, "\n")
	ns := make(map[string]*StorageStatus)

	for _, line := range lines {
		ele := strings.Split(line, "|")
		if len(ele) != 2 {
			//格式不正确
			continue
		}

		nodeName := strings.TrimSpace(ele[0])
		status := strings.TrimSpace(ele[1])

		if strings.HasPrefix(nodeName, "h_") {
			nodeName = nodeName[2:]
		}

		sStatus := &StorageStatus{}
		if status == "online" {
			sStatus.Available = true
			sStatus.Status = status
		} else if status == "degraded" {
			sStatus.Available = true
			sStatus.Status = status
		} else if status == "offline" {
			sStatus.Available = false
			sStatus.Status = status
		} else {
			sStatus.Available = false
			sStatus.Status = status
		}

		ns[nodeName] = sStatus
	}

	return &ns
}

func _GetNodeAddress(node *v1.Node, tagName string) string {
	addressList := node.Status.Addresses
	var nodeAddress = ""
	for _, adr := range addressList {
		if adr.Type == v1.NodeHostName {
			nodeAddress = adr.Address
			break
		}
	}

	if nodeAddress == "" {
		klog.Errorf("[%s] node [%s] node Address is nil [%s], use name instead! ", tagName, node.Name, nodeAddress)
		//return false
		nodeAddress = node.Name
	}
	return nodeAddress
}

func (probe *PolarNodeNetworkProbe) nodeClientAvailable(tagName string) (bool, string, string) {
	checkNicStatusCmd := fmt.Sprintf("ip a show %s|grep \" state \"|grep -e \"%s:\\|%s@\" ", probe.ClientCardName, probe.ClientCardName, probe.ClientCardName)
	if !probe.nodeSshConn.IsInit() {
		iErr := probe.nodeSshConn.Init()
		if iErr != nil {
			return false, "StateUnKnown", probe.ClientCardName + " StateUnKnown"
		}
	}

	stdOut, errInfo, err := probe.nodeSshConn.RunCmdWithLogLevel(
		checkNicStatusCmd, false, int(formatLogLevel(probe.cnt, 5)))

	if err != nil || errInfo != "" && !strings.Contains(errInfo, "Warning: Permanently") {
		if util.GetShellExitCode(err.Error()) == "1" {
			//正常，未查询到任数据何，忽略打印信息
		} else {
			klog.Errorf("[%s] node [%s] card [%s]  check is err[%v] and still check output", tagName, probe.NodeName, probe.ClientCardName, err)
		}
	}

	if strings.Contains(stdOut, "state UP") {
		return true, "StateUP", probe.ClientCardName + " StateUP"
	}

	if strings.Contains(stdOut, "state DOWN") {
		return false, "StateDown", probe.ClientCardName + " StateDown"
	}

	klog.Infof("[%s] node [%s] card [%s]  check is unknown", tagName, probe.NodeName, probe.ClientCardName)

	return false, "StateUnKnown", probe.ClientCardName + " StateUnKnown"
}

func (probe *PolarNodeNetworkProbe) GetHybridDeploySetting() *HybridDeploySetting {
	once.Do(func() {
		if hybridDeploySetting == nil {
			isCheckOObIP, disableSanCmd, err := probe.getHybridDeploy()

			setting := HybridDeploySetting{
				IsCheckOObIP:  isCheckOObIP,
				DisableSanCmd: disableSanCmd,
				Err:           err,
			}
			hybridDeploySetting = &setting
		}
	})

	return hybridDeploySetting
}

func (probe *PolarNodeNetworkProbe) getHybridDeploy() (bool, bool, error) {
	configMapName := "controller-config"
	controllerConfigMap, cmErr := probe.KubeClient.CoreV1().ConfigMaps("kube-system").Get(configMapName, v12.GetOptions{})
	if cmErr != nil {
		return true, false, cmErr
	}

	isCheckOObIPKey := "isCheckOObIP"
	var isCheckOObIP = true
	isCheckOObIPKeyStr, ok := controllerConfigMap.Data[isCheckOObIPKey]
	if !ok {
		klog.Infof("no %v item in cm %s data, use default:true", isCheckOObIPKeyStr, configMapName)
		return true, false, nil
	} else {
		var parseErr error
		isCheckOObIP, parseErr = strconv.ParseBool(isCheckOObIPKeyStr)
		if parseErr != nil {
			klog.Infof("value of item %v in cm %s data is %v. it can't be parsed into bool, parserErr:%v, use default:true",
				isCheckOObIPKey, configMapName, isCheckOObIPKeyStr, parseErr)
		}
	}

	disableSanCmdKey := "disableSanCmd"
	disableSanCmdStr, ok := controllerConfigMap.Data[disableSanCmdKey]
	if !ok {
		klog.Infof("no %v item in cm %s data, default:false", disableSanCmdKey, configMapName)
		return isCheckOObIP, false, nil
	}

	disableSanCmd, err := strconv.ParseBool(disableSanCmdStr)
	if err != nil {
		klog.Infof("value of item %v in cm %s data is %v. it can't be parsed into bool, parserErr:%v, use default:true",
			disableSanCmdKey, configMapName, disableSanCmdStr, err)
		return isCheckOObIP, false, nil
	}

	return isCheckOObIP, disableSanCmd, nil
}
