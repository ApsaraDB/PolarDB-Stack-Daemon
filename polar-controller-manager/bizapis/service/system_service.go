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

package service

import (
	"fmt"
	config "github.com/ApsaraDB/PolarDB-Stack-Daemon/cmd/daemon/app/config"
	"github.com/ApsaraDB/PolarDB-Stack-Daemon/polar-controller-manager/bizapis/context"
	"github.com/ApsaraDB/PolarDB-Stack-Daemon/polar-controller-manager/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

type SystemService struct {
	Config    *config.CompletedConfig
	K8sClient kubernetes.Interface
}

func NewSystemService(
	cfg *config.CompletedConfig, client kubernetes.Interface) *SystemService {
	return &SystemService{
		cfg,
		client}
}

func (s *SystemService) GetPodStandByIp(ctx *context.Context, namespace, podName string) (standbyIp string, err error) {
	if s.K8sClient != nil {
		pod, err := s.K8sClient.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
		if err == nil {
			nodeName := pod.Spec.NodeName
			standbyIp, err := getStandByIpFromNode(s.K8sClient, nodeName)
			if err != nil {
				return standbyIp, err
			} else {
				return standbyIp, nil
			}
		} else {
			klog.Warningf("failed to get pod %q in ns %q, err:%v", podName, namespace, err)
			err := fmt.Errorf("failed to get pod %q in ns %q, err:%v", podName, namespace, err)
			return standbyIp, err
		}
	} else {
		err = fmt.Errorf("failed to k8sClient")
		return standbyIp, err
	}
}

func getStandByIpFromNode(kubeClient kubernetes.Interface, nodeName string) (
	standbyIp string, err error) {

	node, err := kubeClient.CoreV1().Nodes().Get(nodeName, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("Failed to get node %s,  err:%v", nodeName, err)
		return "", err
	}

	standbyIp, err = util.GetNodeConditionValue(node, util.NodeStandbyIP)
	if err != nil {
		klog.Errorf("Failed to get standBy ip of node %s: %v", node.Name, err)
		return "", err
	} else {
		return standbyIp, nil
	}
}
