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

import v1 "k8s.io/api/core/v1"

const (
	SSHRemoteTipsWarningMsgHead = "Warning: Permanently"
	PathPrefix                  = "/api/v1"
	ContextKeyPrefix            = "polar-controller-api"
	ContextKeyResBody           = ContextKeyPrefix + "/res_body"
	TimeFormatMilli             = "2006-01-02T15:04:05.000Z07:00"

	// NodeStandbyIP means the ip is used in standby connection.
	NodeStandbyIP v1.NodeConditionType = "StandbyIP"

	ControllerConfig    = "controller-config"
	CcmConfig           = "ccm-config"
	KubeSystemNamespace = "kube-system"
)
