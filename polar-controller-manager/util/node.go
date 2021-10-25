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
	corev1 "k8s.io/api/core/v1"
	"net"
)

func GetNodeConditionValue(node *corev1.Node, conType corev1.NodeConditionType) (string, error) {
	condition := getNodeCondition(node, conType)
	if condition == nil {
		err := fmt.Errorf("failed to find condition[%v] on node [%v]", conType, node.Name)
		return "", err
	} else if condition.Status == corev1.ConditionFalse {
		err := fmt.Errorf("conditon[%v] on node [%v] is false status, message of cond is %q",
			conType, node.Name, condition.Message)
		return "", err
	} else {
		address := net.ParseIP(condition.Message)
		if address == nil {
			err := fmt.Errorf("message of conditon[%v] on node [%v] is %q, it is invalid ip address",
				conType, node.Name, condition.Message)
			return "", err
		} else {
			return condition.Message, nil
		}
	}
}

func getNodeCondition(node *corev1.Node, conType corev1.NodeConditionType) *corev1.NodeCondition {
	for _, c := range node.Status.Conditions {
		if c.Type == conType {
			return &c
		}
	}
	return nil
}
