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


package util

import (
	"fmt"
	uuid "github.com/satori/go.uuid"
	"strings"
	"sync"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// localService is a local cache try to record the max resource version of each service.
// this is a workaround of BUG #https://github.com/kubernetes/kubernetes/issues/59084
var (
	versionCache *localService
	once         sync.Once
)

type localService struct {
	maxResourceVersion map[string]bool
	lock               sync.RWMutex
}

func (s *localService) set(serviceUID string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.maxResourceVersion[serviceUID] = true
}

func (s *localService) get(serviceUID string) (found bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	_, found = s.maxResourceVersion[serviceUID]
	return
}

func GetKubeClient() (*kubernetes.Clientset, *rest.Config, error) {
	conf, err := config.GetConfig()
	if err != nil {
		klog.Error(err, "unable to set up client config")
		return nil, nil, fmt.Errorf("could not get kubernetes config from kubeconfig: %v", err)
	}

	cs, err := kubernetes.NewForConfig(conf)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not get kubernetes client: %s")
	}
	return cs, conf, nil
}

func GetCoreV1Client() (*corev1.CoreV1Client, error) {
	_, config, err := GetKubeClient()
	if err != nil {
		return nil, errors.Wrap(err, "GetCoreV1Client error")
	}

	client, err := corev1.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "corev1.NewForConfig error")
	}

	return client, nil

}

func GetPodsByLabels(label string, nameSpace string) (*v1.PodList, error) {
	client, _ := GetCoreV1Client()
	pods, err := client.Pods(nameSpace).List(metav1.ListOptions{
		LabelSelector: label,
	})
	if err != nil {
		klog.Infoln(err)
		return nil, errors.Wrap(err, fmt.Sprintf("get pod: %v fail", label))
	}
	return pods, nil
}

func ReduceArray(org []string) []string {
	return reduceArray(org)
}

func reduceArray(org []string) []string {
	if org == nil || len(org) < 1 {
		return org
	}
	var newArray []string
	for _, e := range org {
		trimStr := strings.TrimSpace(e)
		if trimStr == "" {
			continue
		}
		newArray = append(newArray, trimStr)
	}
	return newArray
}

// GetUniqueID get uuid
func GetUniqueID() string {
	return uuid.NewV4().String()
}
