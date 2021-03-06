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

package app

import (
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"k8s.io/kubernetes/pkg/controller"
)

var (
	Conf *CompletedConfig
)

// Config is the main context object for the cloud controller manager.
type Config struct {

	// the general kube client
	Client *clientset.Clientset

	// the client only used for leader election
	LeaderElectionClient *clientset.Clientset

	// the rest config for the master
	Kubeconfig *restclient.Config

	// the event sink
	EventRecorder record.EventRecorder

	// ClientBuilder will provide a client for this controller to use
	ClientBuilder controller.ControllerClientBuilder

	// VersionedClient will provide a client for informers
	VersionedClient clientset.Interface

	// SharedInformers gives access to informers for the controller.
	SharedInformers informers.SharedInformerFactory

	CurrentNodeName            string
	DbclusterLogDir            string
	InsFolderOverdueDays       int32
	DevelopMode                bool
	Port                       int32
	EventsEnableUpload         bool   // 是否开启事件上报
	EventsUploadUrl            string // 事件上报地址
	EventsUploadTimeout        int32  // 事件上报超时 单位 秒
	JobEnableDelFIP            bool   // 是否开启 检查任务中 对无对应服务的IP删除
	JobEnableAddFIP            bool   // 是否开启 检查任务中 对缺失IP的添加
	PolarStackDaemonPodLabels  string // polar stack daemon 的 labels 用于准确查找出 polarstack-daemon 的 Pod 以便调用相关 API
	CoreVersionConfigMapLabel  string // core version configMap 的相关 labels
	MpdControllerConfigMapName string // mpd controller configMap Name (polardb4mpd-controller)
	ServiceOwnerDbCluster      string // service Owner db cluster
}

type completedConfig struct {
	*Config
}

// CompletedConfig same as Config, just to swap private object.
type CompletedConfig struct {
	// Embed a private pointer that cannot be instantiated outside of this package.
	*completedConfig
}

// Complete fills in any fields not set that are required to have valid data. It's mutating the receiver.
func (c *Config) Complete() *CompletedConfig {
	cc := completedConfig{c}

	return &CompletedConfig{&cc}
}
