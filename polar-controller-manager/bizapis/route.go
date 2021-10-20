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


package bizapis

import (
	"github.com/gin-gonic/gin"
	"github.com/ApsaraDB/PolarDB-Stack-Daemon/polar-controller-manager/bizapis/context"
	"github.com/ApsaraDB/PolarDB-Stack-Daemon/polar-controller-manager/util"
)

// const ...
const (
	PublicAPI  = true
	PrivateAPI = false
)

// HandlerFunc 处理函数
type HandlerFunc func(*context.Context)

// Handle registers a new request handle
func Handle(g *gin.RouterGroup, httpMethod string, relativePath string, handler HandlerFunc, isOpen bool, comment string) {

	g.Handle(httpMethod, relativePath, func(c *gin.Context) {
		requestID, ok := c.Get("RequestId")
		if !ok {
			requestID = util.GetUniqueID()
		}

		entry := util.Log.WithField("request_id", requestID)
		handler(context.New(c, entry))
	})
}

// GET is a shortcut for router.Handle("GET", path, handle).
func GET(g *gin.RouterGroup, relativePath string, handler HandlerFunc, isOpen bool, comment string) {
	Handle(g, "GET", relativePath, handler, isOpen, comment)
}

// POST is a shortcut for router.Handle("POST", path, handle).
func POST(g *gin.RouterGroup, relativePath string, handler HandlerFunc, isOpen bool, comment string) {
	Handle(g, "POST", relativePath, handler, isOpen, comment)
}
