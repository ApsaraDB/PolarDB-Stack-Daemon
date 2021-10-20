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


package options

import (
	config "github.com/ApsaraDB/PolarDB-Stack-Daemon/cmd/daemon/app/config"
	"os"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"k8s.io/kubernetes/pkg/controller"
	// add the kubernetes feature gates
	_ "k8s.io/kubernetes/pkg/features"
)

const (
	// CloudControllerManagerUserAgent is the userAgent name when starting cloud-controller managers.
	CloudControllerManagerUserAgent = "polar-controller-manager"
)

// PolarStackControllerManagerOptions is the main context object for the controller manager.
type PolarStackControllerManagerOptions struct {
	Master     string
	Kubeconfig string

	// NodeStatusUpdateFrequency is the frequency at which the controller updates nodes' status
	NodeStatusUpdateFrequency metav1.Duration

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
	ServiceOwnerDbCluster      string // service owner db cluster
}

func NewPolarStackControllerManagerOptions() (*PolarStackControllerManagerOptions, error) {

	s := PolarStackControllerManagerOptions{}

	return &s, nil
}

// Config return a cloud controller manager config objective
func (o *PolarStackControllerManagerOptions) Config() (*config.Config, error) {
	c := &config.Config{}
	if err := o.ApplyTo(c, CloudControllerManagerUserAgent); err != nil {
		return nil, err
	}

	return c, nil
}

// Flags returns flags for a specific APIServer by section name
func (o *PolarStackControllerManagerOptions) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}

	fs := fss.FlagSet("misc")
	fs.StringVar(&o.Master, "master", o.Master, "The address of the Kubernetes API server (overrides any value in kubeconfig).")
	fs.StringVar(&o.Kubeconfig, "kubeconfig", o.Kubeconfig, "Path to kubeconfig file with authorization and master location information.")
	fs.DurationVar(&o.NodeStatusUpdateFrequency.Duration, "node-status-update-frequency", o.NodeStatusUpdateFrequency.Duration, "Specifies how often the controller updates nodes' status.")
	fs.StringVar(&o.CurrentNodeName, "current-node-name", os.Getenv("CURRENT_NODE_NAME"), "current node name")
	fs.StringVar(&o.DbclusterLogDir, "dbcluster-log-dir", "/flash/polardb_dbcluster/", "dbcluster log path in host")
	fs.Int32Var(&o.InsFolderOverdueDays, "ins-folder-overdue-days", 3, "deleted instance folder overdue days")
	fs.BoolVar(&o.DevelopMode, "develop-mode", false, "develop-mode")
	fs.Int32Var(&o.Port, "port", 8900, "http server port")
	fs.BoolVar(&o.EventsEnableUpload, "events-enable-upload", false, "is events upload enabled")
	fs.StringVar(&o.EventsUploadUrl, "events-upload-url", "http://rds-redline-worker-log-api.rds:8080/", "event upload url")
	fs.Int32Var(&o.EventsUploadTimeout, "events-upload-timeout", 3, "events upload timeout")
	fs.BoolVar(&o.JobEnableDelFIP, "job-enable-del-fip", true, "job enable del floating ip")
	fs.BoolVar(&o.JobEnableAddFIP, "job-enable-add-fip", true, "job enable add floating ip")
	fs.StringVar(&o.PolarStackDaemonPodLabels, "polarstack-daemon-pod-labels", "app=polarstack-daemon", "polar stack daemon pod labels")
	fs.StringVar(&o.CoreVersionConfigMapLabel, "core-version-cm-labels", "configtype=minor_version_info,dbClusterMode=WriteReadMore", "core version configMap labels")
	fs.StringVar(&o.MpdControllerConfigMapName, "mpd-controller-cm-name", "polardb4mpd-controller", "mpd controller configMap name ")
	fs.StringVar(&o.ServiceOwnerDbCluster, "service-owner-db-cluster", "mpdcluster", "service owner db cluster")
	return fss
}

// ApplyTo fills up cloud controller manager config with options.
func (o *PolarStackControllerManagerOptions) ApplyTo(c *config.Config, userAgent string) error {
	var err error

	c.Kubeconfig, err = clientcmd.BuildConfigFromFlags(o.Master, o.Kubeconfig)
	if err != nil {
		return err
	}
	c.Client, err = clientset.NewForConfig(restclient.AddUserAgent(c.Kubeconfig, userAgent))
	if err != nil {
		return err
	}

	c.EventRecorder = createRecorder(c.Client, userAgent)
	rootClientBuilder := controller.SimpleControllerClientBuilder{
		ClientConfig: c.Kubeconfig,
	}
	c.ClientBuilder = rootClientBuilder
	c.CurrentNodeName = o.CurrentNodeName
	c.DbclusterLogDir = o.DbclusterLogDir
	c.InsFolderOverdueDays = o.InsFolderOverdueDays
	c.DevelopMode = o.DevelopMode
	c.Port = o.Port
	c.EventsUploadUrl = o.EventsUploadUrl
	c.EventsEnableUpload = o.EventsEnableUpload
	c.EventsUploadTimeout = o.EventsUploadTimeout
	c.JobEnableDelFIP = o.JobEnableDelFIP
	c.JobEnableAddFIP = o.JobEnableAddFIP
	c.PolarStackDaemonPodLabels = o.PolarStackDaemonPodLabels
	c.CoreVersionConfigMapLabel = o.CoreVersionConfigMapLabel
	c.MpdControllerConfigMapName = o.MpdControllerConfigMapName
	c.ServiceOwnerDbCluster = o.ServiceOwnerDbCluster
	return nil
}

func createRecorder(kubeClient clientset.Interface, userAgent string) record.EventRecorder {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})

	return eventBroadcaster.NewRecorder(legacyscheme.Scheme, v1.EventSource{Component: userAgent})
}
