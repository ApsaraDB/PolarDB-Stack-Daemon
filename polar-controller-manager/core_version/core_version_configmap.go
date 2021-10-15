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
	"errors"
	"fmt"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"strings"
)

// getAllCoreVersionConfigMap
/**
 * @Title:  getAllCoreVersionConfigMap
 * @Description: 获取所有内核版本相关的 configMap
 **/
func getAllCoreVersionConfigMapByLabels(client *clientset.Clientset, labels string) (coreVersions *v1.ConfigMapList, err error) {
	coreVersions, err = client.CoreV1().ConfigMaps(NameSpace).List(metav1.ListOptions{LabelSelector: labels})
	return
}

// getAllImagesByVersionConfigMap
/**
 * @Title:  getAllImagesByVersionConfigMap
 * @Description: 从 core version 的 configMap 中获取所有镜像的信息
 * 此类 configMap 中的 Data 中 key 中包括 Image 关键字，以此为依据
 * 并剔除 image 为空的对象
 **/
func getAllImagesByVersionConfigMap(version *v1.ConfigMap) (images []string, err error) {
	if nil == version {
		err = errors.New("version is nil")
		return
	}
	if nil == version.Data {
		err = errors.New("version.Data is nil")
		return
	}
	for key, image := range version.Data {
		if strings.Index(key, coreVersionImageKeyKeyWord) > 0 && len(image) > 0 {
			images = append(images, image)
		}
	}
	return
}

// getCoreVersionNameByVersionConfigMap
/**
 * @Title:  getCoreVersionNameByVersionConfigMap
 * @Description: 从 core version 的 configMap 中获取当前 core 的版本号
 **/
func getCoreVersionNameByVersionConfigMap(version *v1.ConfigMap) (versionName string, err error) {
	if nil == version {
		err = errors.New("version is nil")
		return
	}
	if nil == version.Data {
		err = errors.New("version.Data is nil")
		return
	}
	for key, info := range version.Data {
		if key == coreVersionNameKey && len(info) > 0 {
			versionName = info
			break
		}
	}
	return
}

// getHostCoreVersionConfigMap
/**
 * @Title:  getHostCoreVersionConfigMap
 * @Description: 获取主机 core version 的 configMap
 **/
func getHostCoreVersionConfigMap(client *clientset.Clientset) (cm *v1.ConfigMap, err error) {
	configName := fmt.Sprintf(coreVersionAvailabilityConfigName, hostName)
	cm, err = client.CoreV1().ConfigMaps(NameSpace).Get(configName, metav1.GetOptions{})
	if err != nil && apierrors.IsNotFound(err) {
		cm = &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      configName,
				Namespace: NameSpace,
			},
		}
		cm, err = client.CoreV1().ConfigMaps(NameSpace).Create(cm)
		if err != nil {
			klog.Errorf("%s failed to create cm:%s, err:%s", logInfoTarget, configName, err.Error())
			return
		}
	} else if err != nil {
		klog.Errorf("%s failed to get cm:%s, err:%s", logInfoTarget, configName, err.Error())
		return
	}
	if cm.Data == nil {
		cm.Data = make(map[string]string)
	}

	return
}

// updateHostCoreVersionConfigMap
/**
 * @Title:  updateHostCoreVersionConfigMap
 * @Description: 更新主机上 core version 的版本信息
 **/
func updateHostCoreVersionConfigMap(client *clientset.Clientset, cm *v1.ConfigMap) (rcm *v1.ConfigMap, err error) {
	return client.CoreV1().ConfigMaps(NameSpace).Update(cm)
}
