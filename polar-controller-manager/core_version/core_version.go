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


// Package core_version
/**
 * @Title: core version 检查
 * @Description:
 *
 *   core version 检查，含三个部份：
 *   - API：提供外部与内部两个接口调用，触发 core version 的检测，外部调用时将通过内部API通知其它节点开启检测
 *   - 检测体：polar stack daemon 启动时，将对 CheckRequestQueue 进行监听，一旦收到请求，开始检查 core version 镜像是否存在
 *   - 对 k8s configMap 与 镜像的操作
 *    另外，每次启动主动更新一次主机的core version信息
 *
 */
package core_version

import (
	"os"
	"strings"
)

//CheckRequestQueue
/**
 * @Description: 请求检查队列
 **/
var CheckRequestQueue chan OperatorType

type OperatorType string

var NameSpace = "kube-system"
var hostName string

// polar stack daemon 的 labels 用于准确查找出 polarstack-daemon 的 Pod 以便调用相关 API
var polarStackDaemonLabels = "app=polarstack-daemon"

// core version configMap 的相关 labels
var coreVersionConfigMapLabel = "configtype=minor_version_info,dbClusterMode=WriteReadMore"

const (
	// core version configMap 的 data 中 Key 如果包含 Image 则为镜像
	coreVersionImageKeyKeyWord = "Image"
	// core version configMap 的 data 中 Key 如果是 name 则为版本号
	coreVersionNameKey = "name"
	// 主机记录 core version 的 config name 格式，需采用 hostName 替换
	coreVersionAvailabilityConfigName = "polarstack-daemon-version-availability-%s"
	// 统一时间格式
	timeFormat = "2006-01-02T15:04:05Z"
	// 日志标记，所有日志将加上该标记，以便于测试与分析
	logInfoTarget = "[core_version]"
)

// CheckCoreVersionOperatorType 外部请求检查 core version
var CheckCoreVersionOperatorType OperatorType = "CheckCoreVersionOperatorType"

// SingleCheckCoreVersionOperatorType 来自内部的检查通知
var SingleCheckCoreVersionOperatorType OperatorType = "SingleCheckCoreVersionOperatorType"

func (opt OperatorType) ToString() string {
	switch opt {
	case CheckCoreVersionOperatorType:
		return "CheckCoreVersionOperatorType"
	case SingleCheckCoreVersionOperatorType:
		return "SingleCheckCoreVersionOperatorType"
	default:
		return "CheckCoreVersionOperatorType"
	}
}

/**
 * @Title:  core version init
 * @Description: 初始化 core version
 **/
func init() {
	CheckRequestQueue = make(chan OperatorType, 2)
	hostName, _ = os.Hostname()
	// hostName 作为 cm 名称的一部份，需小写
	hostName = strings.ToLower(hostName)
}
