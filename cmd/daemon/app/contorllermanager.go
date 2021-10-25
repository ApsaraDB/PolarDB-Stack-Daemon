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
	"fmt"
	config "github.com/ApsaraDB/PolarDB-Stack-Daemon/cmd/daemon/app/config"
	"github.com/ApsaraDB/PolarDB-Stack-Daemon/cmd/daemon/app/options"
	"github.com/ApsaraDB/PolarDB-Stack-Daemon/polar-controller-manager/bizapis"
	"github.com/ApsaraDB/PolarDB-Stack-Daemon/polar-controller-manager/core_version"
	"github.com/ApsaraDB/PolarDB-Stack-Daemon/polar-controller-manager/db_log_monitor"
	"github.com/ApsaraDB/PolarDB-Stack-Daemon/polar-controller-manager/events"
	"github.com/ApsaraDB/PolarDB-Stack-Daemon/polar-controller-manager/node_net_status"
	usage "github.com/ApsaraDB/PolarDB-Stack-Daemon/polar-controller-manager/port_usage"
	cmversion "github.com/ApsaraDB/PolarDB-Stack-Daemon/version"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/component-base/cli/globalflag"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/version/verflag"
	"os"
)

func NewControllerManagerCommand() *cobra.Command {
	s, err := options.NewPolarStackControllerManagerOptions()
	if err != nil {
		klog.Fatalf("unable to initialize command options: %v", err)
	}

	cmd := &cobra.Command{
		Use: "polar-controller-manager",
		Long: `The Cloud controller manager is a daemon that embeds
the cloud specific control loops shipped with Kubernetes.`,
		Run: func(cmd *cobra.Command, args []string) {

			klog.Infof("CODE_SOURCE: %s", os.Getenv("CODE_SOURCE"))
			klog.Infof("CODE_BRANCHES: %s", os.Getenv("CODE_BRANCHES"))
			klog.Infof("CODE_VERSION: %s", os.Getenv("CODE_VERSION"))

			klog.Infof("---------------------------------------------------------------------------------------------")
			klog.Infof("|                                                                                           |")
			klog.Infof("| branch:%v commitId:%v \n", cmversion.GitBranch, cmversion.GitCommitId)
			klog.Infof("| polarbox repo %v\n", cmversion.GitCommitRepo)
			klog.Infof("| polarbox commitDate/buildDate %v\n", cmversion.GitCommitDate)
			klog.Infof("|                                                                                           |")
			klog.Infof("---------------------------------------------------------------------------------------------")

			c, err := s.Config()
			klog.Infof("ccm config: %+v", c)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
			config.Conf = c.Complete()
			//
			if err := Run(c.Complete(), wait.NeverStop); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}

		},
	}
	fs := cmd.Flags()
	namedFlagSets := s.Flags()
	verflag.AddFlags(namedFlagSets.FlagSet("global"))
	globalflag.AddGlobalFlags(namedFlagSets.FlagSet("global"), cmd.Name())

	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}
	return cmd
}

// Run runs the ExternalCMServer.  This should never exit.
func Run(completedConfig *config.CompletedConfig, stopCh <-chan struct{}) error {

	client := completedConfig.ClientBuilder.ClientOrDie("polarstack-daemon")

	klog.Info("start timer StartLogMonitor")
	go db_log_monitor.StartLogMonitor(stopCh)
	klog.Info("start timer StartNodeNetworkProbe")
	go node_net_status.StartNodeNetworkProbe(client.(*clientset.Clientset), stopCh)
	klog.Info("start timer StartPrintPort")
	go usage.StartPrintPort(client.(*clientset.Clientset), stopCh)
	klog.Info("start StartCheckCoreVersion")
	go core_version.StartCheckCoreVersion(client.(*clientset.Clientset))

	port := config.Conf.Port
	klog.Infof("webserver port:%d", port)
	bizapis.StartHttpServer(completedConfig, client)

	// 初使化事件上报对象
	events.Init()
	<-stopCh
	panic("unreachable")
}
