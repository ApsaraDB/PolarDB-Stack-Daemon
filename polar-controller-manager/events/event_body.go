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


package events

// EventBody
/**
 * @Title:  上报日志的请求消息体
 **/
type EventBody struct {
	// 事件发生时业务域内唯一ID
	EventId string `json:"eventId"`
	// 事件规则编码
	EventCode EventCode `json:"eventCode"`
	// 来源
	Source Source `json:"source"`
	// 事件的级别
	//  - CRITICAL : "不可逆错误，等同系统宕机"
	//  - ERROR    : "系统错误，流程还能继续"
	//  - WARN     : "不影响系统正常流程"
	//  - INFO     : "提示信息"
	Level EventLevel `json:"level"`
	// 事件发生时间 毫秒级时间戳
	Time int64 `json:"time"`
	// 事件内容，json格式
	Body Body `json:"body"`
}

type Body struct {
	// 描述，告警时会使用
	Describe string `json:"describe"`
}

type Source struct {
	// 资源类型
	ResourceType string `json:"resourceType"`
	// 实例Name
	InsName string `json:"insName"`
	// 发现问题的组件类型，英文编码（各组件自己命名）
	From string `json:"from"`
	// 发生的IP地址，如果拿不到发生的IP地址，可以置空。
	Ip string `json:"ip"`
}
