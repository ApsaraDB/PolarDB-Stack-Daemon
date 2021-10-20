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


package controller

import (
	"fmt"
	gincontext "github.com/ApsaraDB/PolarDB-Stack-Daemon/polar-controller-manager/bizapis/context"
	"github.com/ApsaraDB/PolarDB-Stack-Daemon/polar-controller-manager/bizapis/service"
	"k8s.io/klog"
)

type SystemController struct {
	systemService *service.SystemService
}

func NewSystemController(systemService *service.SystemService) *SystemController {
	return &SystemController{systemService}
}

// GetPodStandByIp
/**
获取node指定condition字段内容，用于standby， 这里是示例内容，暂时未在node上提供standby ip信息
*/
func (s *SystemController) GetPodStandByIp(ctx *gincontext.Context) {
	namespace := ctx.GetContext().DefaultQuery("namespace", "default")
	clusterId := ctx.GetContext().Query("clusterId")
	podName := ctx.GetContext().Query("podName")
	dbClusterType := ctx.GetContext().Query("type")
	klog.Infof("GetStandByIp namespace:%s, clusterId:%s, podName:%s, dbClusterType:%s", namespace, clusterId, podName, dbClusterType)
	if podName == "" {
		err := fmt.Errorf("failed for podName is empty in url query parameter")
		ctx.ResErr(err)
	} else {
		resp, err := s.systemService.GetPodStandByIp(ctx, namespace, podName)
		if err != nil {
			klog.Errorf("failed to get standBy ip for pod %s, err:%v", podName, err)
			ctx.ResErr(err)
			return
		}
		ctx.ResSucData(resp)
	}

}
