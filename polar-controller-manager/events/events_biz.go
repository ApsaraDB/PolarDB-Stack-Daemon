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


// Package events
/**
 * @Title:  事件业务处理
 * @Description:
 *
 *	事件业务处理
 *
 **/
package events

import (
	"errors"
	"fmt"
	config "github.com/ApsaraDB/PolarDB-Stack-Daemon/cmd/daemon/app/config"
	"github.com/ApsaraDB/PolarDB-Stack-Daemon/polar-controller-manager/util"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/klog"
	"time"
)

var httpClient *util.HttpClient

// eventsEnableUpload 是否开始上报
var eventsEnableUpload bool

// EventLevel 事件级别
type EventLevel string

// ToString
/**
 * @Title:  EventLevel.ToString
 * @Description:
 *
 *	将 EventLevel 由枚举类型转字符型，默认为 "INFO"
 *
 **/
func (el EventLevel) ToString() string {
	switch el {
	case EventLevelInfo:
		return "INFO"
	case EventLevelWarn:
		return "WARN"
	case EventLevelError:
		return "ERROR"
	case EventLevelCritical:
		return "CRITICAL"
	default:
		return "INFO"
	}
}

// EventCode 事件代码
type EventCode string

// ToString
/**
 * @Title:  EventCode.ToString
 * @Description:
 *
 *	将事件代码由枚举类型转字符型
 *
 **/
func (ec EventCode) ToString() string {
	switch ec {
	case EventFloatingIPAddSuccess:
		return "FloatingIPAddSuccess"
	case EventFloatingIPAddFailed:
		return "FloatingIPAddFailed"
	case EventFloatingIPDeleteSuccess:
		return "FloatingIPDeleteSuccess"
	case EventFloatingIPDeleteFailed:
		return "FloatingIPDeleteFailed"
	case EventNeedToAddFloatingIP:
		return "EventNeedToAddFloatingIP"
	case EventRemoveUnUseFloatingIPSuccess:
		return "EventRemoveUnUseFloatingIPSuccess"
	case EventRemoveUnUseFloatingIPFailed:
		return "EventRemoveUnUseFloatingIPFailed"
	default:
		return "Unknown"
	}
}

const (
	// EventResourceType 事件上报的资源类型
	EventResourceType = "polardb-o"
	// EventSourceFrom 事件上报的资源类型
	EventSourceFrom = "Polardbstack-Daemon"
	// EventHeader1 事件上报请求头部信息 1
	EventHeader1 = "event_resource;log;20210220001;POLAR_STACK"
	// EventHeader2 事件上报请求头部信息 2
	EventHeader2 = "event"

	// EventLevelInfo 提示信息
	EventLevelInfo EventLevel = "INFO"
	// EventLevelWarn 不影响系统正常流程
	EventLevelWarn EventLevel = "WARN"
	// EventLevelError 系统错误，流程还能继续
	EventLevelError EventLevel = "ERROR"
	// EventLevelCritical 不可逆错误，等同系统宕机
	EventLevelCritical EventLevel = "CRITICAL"

	// EventFloatingIPAddSuccess 浮动IP添加成功
	EventFloatingIPAddSuccess EventCode = "FloatingIPAddSuccess"
	// EventFloatingIPAddFailed 浮动IP添加失败
	EventFloatingIPAddFailed EventCode = "FloatingIPAddFailed"
	// EventFloatingIPDeleteSuccess 浮动IP删除成功
	EventFloatingIPDeleteSuccess EventCode = "FloatingIPDeleteSuccess"
	// EventFloatingIPDeleteFailed 浮动IP删除失败
	EventFloatingIPDeleteFailed EventCode = "FloatingIPDeleteFailed"

	// EventNeedToAddFloatingIP 应该存在该ip， 但却不存在 需要添加
	EventNeedToAddFloatingIP EventCode = "EventNeedToAddFloatingIP"
	// EventRemoveUnUseFloatingIPSuccess 删除不应该存在的IP成功
	EventRemoveUnUseFloatingIPSuccess EventCode = "EventRemoveUnUseFloatingIPSuccess"
	// EventRemoveUnUseFloatingIPFailed 删除不应该存在的IP失败
	EventRemoveUnUseFloatingIPFailed EventCode = "EventRemoveUnUseFloatingIPFailed"
)

// Init
/**
 * @Title:  初使化事件上报对象
 * @Description:
 *
 * 初使化事件上报的请求对象及参数：
 *    - 上报地址
 *    - 上报超时时间
 *    - 是否开启事件上报
 **/
func Init() {
	httpClient = &util.HttpClient{Host: config.Conf.EventsUploadUrl, Timeout: time.Duration(config.Conf.EventsUploadTimeout) * time.Second}
	eventsEnableUpload = config.Conf.EventsEnableUpload
}

// UploadEvent
/**
 * @Title:  事件上报
 **/
func UploadEvent(eventCode EventCode, insName string, ip string, describe string) (er EventResponse, err error) {
	if !eventsEnableUpload {
		err = errors.New("UploadEvent: events upload disabled")
		return
	}

	eventBody := EventBody{}
	eventBody.Body.Describe = describe
	eventBody.EventCode = eventCode
	eventBody.Source.ResourceType = EventResourceType
	eventBody.Source.From = EventSourceFrom
	eventBody.Source.Ip = ip
	eventBody.Source.InsName = insName

	eventBody.Level = getEventLevel(eventCode)
	eventBody.EventId = fmt.Sprintf("%s-%s", eventBody.Source.From, uuid.NewUUID())
	klog.Infof("UploadEvent eventBody info: %#v", eventBody)

	eventBody.Time = time.Now().UnixNano() / 1000000

	path := "/data/upload"
	header := make(map[string]string)
	header["header1"] = EventHeader1
	header["header2"] = EventHeader2
	header["identityKey"] = fmt.Sprintf("%s;%s;", insName, ip)

	resp, err := httpClient.Post(path, header, eventBody)
	if resp != nil && resp.Body != nil {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			klog.Infof("failed to close httpResp, err:%v", closeErr)
		}
	}
	if err != nil {
		klog.Infof("UploadEvent failed err:%v，statusCode:%v", err, resp)
	}
	return
}

// getEventLevel
/**
 * @Title:  获取事件代码对应的事件等级
 **/
func getEventLevel(eventCode EventCode) EventLevel {
	// 事件代码与等级匹配
	switch eventCode {
	case EventFloatingIPAddSuccess:
		return EventLevelInfo
	case EventFloatingIPDeleteSuccess:
		return EventLevelInfo
	case EventFloatingIPAddFailed:
		return EventLevelWarn
	case EventFloatingIPDeleteFailed:
		return EventLevelWarn
	case EventRemoveUnUseFloatingIPSuccess:
		return EventLevelWarn
	case EventRemoveUnUseFloatingIPFailed:
		return EventLevelWarn
	case EventNeedToAddFloatingIP:
		return EventLevelWarn
	default:
		return EventLevelInfo
	}
}
