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

package context

import (
	"encoding/json"
	base "github.com/ApsaraDB/PolarDB-Stack-Daemon/polar-controller-manager/bizapis/model"
	"github.com/ApsaraDB/PolarDB-Stack-Daemon/polar-controller-manager/errors"
	"github.com/ApsaraDB/PolarDB-Stack-Daemon/polar-controller-manager/util"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

// Context Context
type Context struct {
	gctx *gin.Context
	Log  *logrus.Entry
}

// New New
func New(c *gin.Context, log *logrus.Entry) *Context {
	return &Context{
		gctx: c,
		Log:  log,
	}
}

// GetContext GetContext
func (a *Context) GetContext() *gin.Context {
	return a.gctx
}

// GetRawData Get gin context raw data
func (a *Context) GetRawData() ([]byte, error) {
	return a.gctx.GetRawData()
}

// ResSucData ResSucData
func (a *Context) ResSucData(data interface{}) {
	resp := base.BaseResponseResult{
		RequestId: a.gctx.GetString("RequestId"),
		Data:      data,
		Code:      "1",
	}
	a.ResJSON(http.StatusOK, resp)
}

// ResErr ResErr
func (a *Context) ResErr(err error, data ...interface{}) {
	a.ResError(err, "", data...)
}

// ResError ResSuccess body has err msg
func (a *Context) ResError(err error, respCode string, data ...interface{}) {
	if err == nil {
		panic("err can't be null")
	}

	var (
		httpCode int
		resp     = base.BaseResponseResult{
			RequestId: a.gctx.GetString("RequestId"),
			Message:   err.Error(),
		}
	)

	if len(data) != 0 {
		resp.Data = data[0]
	}

	switch e := err.(type) {
	case *errors.MessageError:
		resp.Code = e.Code()
		httpCode = getStatusByError(e.Parent())
	default:
		resp.Code = strconv.Itoa(httpCode)
		httpCode = getStatusByError(err)
	}

	if respCode != "" {
		resp.Code = respCode
	}

	a.ResJSON(httpCode, resp)
}

// 根据错误获取状态码
func getStatusByError(err error) (status int) {
	switch err {
	case errors.ErrNormal:
		status = 200
	case errors.ErrBadRequest:
		status = 400
	case errors.ErrUnauthorized:
		status = 401
	case errors.ErrForbidden:
		status = 403
	case errors.ErrNotFound:
		status = 404
	case errors.ErrInternalServer:
		status = 500
	case errors.ErrImplemented:
		status = 501
	default:
		status = 500
	}
	return
}

// ResJSON 响应JSON数据
func (a *Context) ResJSON(status int, v interface{}) {
	buf, err := json.Marshal(v)
	if err != nil {
		a.ResErr(err)
		return
	}
	a.gctx.Set(util.ContextKeyResBody, buf)
	a.gctx.Data(status, "application/json; charset=utf-8", buf)
	a.gctx.Abort()
}
