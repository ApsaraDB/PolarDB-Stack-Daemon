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


package alicloud

import (
	"fmt"
	"gitlab.alibaba-inc.com/rds/polarstack-daemon/polar-controller-manager/util"
	"strings"
	"time"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

type ControllerConfig struct {
	SshUser         string
	EnablePrintPort bool
}

type GetControllerConfigPolicy struct {
	IsInited         bool
	LastQueryTime    *time.Time
	MaxQueryStepTime *time.Duration    //如果值为0表示只读一次,此值设置原则上不允许修改，否则会逻辑错乱
	LastValue        *ControllerConfig //最后一次结果暂存器
}

var ControllerConfigPolicy *GetControllerConfigPolicy

func init() {
	maxQueryStepTime := 30 * time.Minute
	ControllerConfigPolicy = &GetControllerConfigPolicy{
		IsInited:         false,
		MaxQueryStepTime: &maxQueryStepTime,
	}
}

func CheckPolicyIsReload(policy *GetControllerConfigPolicy) bool {
	if policy == nil || !policy.IsInited || policy.LastQueryTime == nil || policy.LastValue == nil {
		return true
	}

	//未设置允许时长，只读一次
	if policy.MaxQueryStepTime == nil || policy.MaxQueryStepTime.Seconds() == 0 {
		return false
	}

	stepSeconds := policy.MaxQueryStepTime.Seconds()
	realSeconds := time.Now().Sub(*policy.LastQueryTime).Seconds()
	if realSeconds >= stepSeconds {
		return true
	}
	return false
}

func GetControllerConfig() (*ControllerConfig, error) {
	return GetControllerConfigWithPolicy(ControllerConfigPolicy)
}

func GetControllerConfigWithPolicy(policy *GetControllerConfigPolicy) (*ControllerConfig, error) {
	if policy != nil {
		reload := CheckPolicyIsReload(policy)
		if !reload {
			return policy.LastValue, nil
		}
	}
	klog.Infof("reload cm %s", util.ControllerConfig)
	client, _ := util.GetCoreV1Client()
	conf, err := client.ConfigMaps(util.KubeSystemNamespace).Get(util.ControllerConfig, metav1.GetOptions{})
	if err != nil {
		err = errors.Errorf("failed to get cm %q, error:%v", util.ControllerConfig, err)
		klog.Error(err)
		return nil, err
	}

	var result = &ControllerConfig{}

	sshUser, err := parseMapItemToString(conf.Data, "sshUser")
	if err != nil {
		return nil, err
	}
	result.SshUser = sshUser

	enablePrintPort, err := parseMapItemToBoolWithDefault(conf.Data, "enablePrintPort", true)
	if err != nil {
		return nil, err
	}
	result.EnablePrintPort = enablePrintPort

	if policy != nil {
		policy.LastValue = result
		policy.IsInited = true
		now := time.Now()
		policy.LastQueryTime = &now
	}

	return result, nil
}

func parseMapItemToBoolWithDefault(data map[string]string, key string, defaultValue bool) (bool, error) {
	valueStr, ok := data[key]
	if !ok {
		logData := maskMapDataPwdForLog(data)
		klog.Errorf("key [%s] not found from map %v", key, logData)
	}
	if valueStr == "" {
		return defaultValue, nil
	}
	strTrim := strings.ToUpper(strings.TrimSpace(valueStr))
	if "1" == strTrim || "T" == strTrim || "Y" == strTrim || "TRUE" == strTrim || "YES" == strTrim {
		return true, nil
	} else {
		if defaultValue {
			//默认为true，预期之外的字符串应视为true，即言下之意，False也必须在枚举范围之内
			return "0" != strTrim && "F" != strTrim && "N" != strTrim && "FALSE" != strTrim && "NO" != strTrim, nil
		} else {
			//默认为true ,预期之外的字符串应视为false
			return false, nil
		}
	}
}

func parseMapItemToString(data map[string]string, key string) (string, error) {
	value, ok := data[key]
	if !ok {
		err := errors.Errorf("key [%s] not found from map %v", key, data)
		klog.Error(err)
		return "", err
	}
	return value, nil
}

func maskMapDataPwdForLog(data map[string]string) map[string]string {
	if data == nil {
		return nil
	}
	ret := make(map[string]string)
	for k, v := range data {
		if k == "sshPassword" {

			if len(v) < 1 {
				ret[k] = "[nopwd]"
				continue
			}

			subStrNum := 2
			if len(v) < 2 {
				subStrNum = 1
			}

			head := v[:subStrNum]
			tail := v[len(v)-subStrNum:]
			nv := fmt.Sprintf("[%s******%s]", head, tail)
			ret[k] = nv
		} else {
			ret[k] = v
		}
	}
	return ret
}
