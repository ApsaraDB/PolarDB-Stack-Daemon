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
	"context"
	. "docker.io/go-docker"
	"k8s.io/klog"
)

// 每次检测镜像是否存在，将结果临时 cache 住，在同一批次的检查中，避免重复。
var cache map[string]bool

// ClearImagesCache
/**
 * @Title:  ClearImagesCache
 * @Description:
 *
 *	清空 镜像检测的缓存
 **/
func ClearImagesCache() {
	cache = make(map[string]bool, 0)
}

// ImageIsExists
/**
 * @Title:  ImageIsExists
 * @Description:
 *
 *	通过 docker sdk api 检查镜像是否存在于本机
 *
 **/
func ImageIsExists(image string, logPrefix string) (exists bool) {
	var client *Client
	defer func() {
		if err := recover(); err != nil {
			klog.Errorf("%s failed to check docker image:%s, err:%v", logPrefix, image, err)
		}
		if client != nil {
			err := client.Close()
			if err != nil {
				klog.Errorf("%s failed to close docker client when checking image:%s, err:%v", logPrefix, image, err)
			}
		}
	}()
	status, exists := cache[image]
	if exists {
		return status
	}

	client, err := NewEnvClient()
	if err != nil {
		klog.Warningf("%s failed to create docker client to check image, return false by default. err:%v",
			logPrefix, err)
		return false
	}
	klog.Infof("Current host docker client version: %#v", client.ClientVersion())

	imageInspect, _, err := client.ImageInspectWithRaw(context.Background(), image)
	if err != nil {
		klog.Errorf("%s failed to docker inspect image:%s, err:%s", logPrefix, image, err.Error())
		cache[image] = false
		return false
	} else {
		klog.Infof("%s found image:%s, id:%v", logPrefix, image, imageInspect.ID)
	}

	cache[image] = true
	return true
}
