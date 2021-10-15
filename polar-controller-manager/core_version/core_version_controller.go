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
	"gitlab.alibaba-inc.com/rds/polarstack-daemon/polar-controller-manager/bizapis/context"
)

// RequestCheckCoreVersion
/**
 * @Title:  RequestCheckCoreVersion
 * @Description: 外部调用，通知 polar stack daemon 检查 core version，收到通知后将会通知其它 polarstack-daemon 节点
 **/
func RequestCheckCoreVersion(ctx *context.Context) {
	CheckRequestQueue <- CheckCoreVersionOperatorType
	ctx.ResSucData("done")
}

// InnerCheckCoreVersion
/**
 * @Title:  InnerCheckCoreVersion
 * @Description: 由内部触发，通知当前 polarstack daemon 检查 core version，将不再通知其它节点
 **/
func InnerCheckCoreVersion(ctx *context.Context) {
	b, _ := ctx.GetRawData()
	ctx.Log.Infof("%s current host is : %s. request comes from %s", logInfoTarget, hostName, string(b))
	CheckRequestQueue <- SingleCheckCoreVersionOperatorType
	ctx.ResSucData("OK")
}
