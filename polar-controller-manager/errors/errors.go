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

package errors

import (
	"github.com/pkg/errors"
)

// 定义通用错误
var (
	New               = errors.New
	Wrap              = errors.Wrap
	Wrapf             = errors.Wrapf
	ErrNormal         = New("返回code为200，特殊异常")
	ErrForbidden      = New("禁止访问")
	ErrNotFound       = New("资源不存在")
	ErrBadRequest     = New("请求无效")
	ErrUnauthorized   = New("未授权")
	ErrInternalServer = New("服务器错误")
	ErrImplemented    = New("API未实现触发异常")
)

// NewBadRequestError 创建请求无效错误
func NewBadRequestError(code string, msg ...string) error {
	return NewMessageError(ErrBadRequest, code, msg...)
}

// NewUnauthorizedError 创建未授权错误
func NewUnauthorizedError(code string, msg ...string) error {
	return NewMessageError(ErrUnauthorized, code, msg...)
}

// NewForbiddenError 创建资源禁止访问错误
func NewForbiddenError(code string, msg ...string) error {
	return NewMessageError(ErrForbidden, code, msg...)
}

// NewNotFoundError 创建资源不存在错误
func NewNotFoundError(code string, msg ...string) error {
	return NewMessageError(ErrNotFound, code, msg...)
}

// NewInternalServerError 创建服务器错误
func NewInternalServerError(code string, msg ...string) error {
	return NewMessageError(ErrInternalServer, code, msg...)
}

func NewNormalError(code string, msg ...string) error {
	return NewMessageError(ErrNormal, code, msg...)
}

// NewMessageError 创建自定义消息错误
func NewMessageError(parent error, code string, msg ...string) error {
	if parent == nil {
		return nil
	}

	m := parent.Error()
	if len(msg) > 0 {
		m = msg[0]
	}
	return &MessageError{parent, m, code}
}

// MessageError 自定义消息错误
type MessageError struct {
	err  error
	msg  string
	code string
}

func (m *MessageError) Error() string {
	return m.msg
}

func (m *MessageError) Code() string {
	return m.code
}

// Parent 父级错误
func (m *MessageError) Parent() error {
	return m.err
}
