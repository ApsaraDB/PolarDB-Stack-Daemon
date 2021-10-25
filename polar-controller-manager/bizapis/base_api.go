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
	gincontext "github.com/ApsaraDB/PolarDB-Stack-Daemon/polar-controller-manager/bizapis/context"
	"github.com/ApsaraDB/PolarDB-Stack-Daemon/polar-controller-manager/bizapis/model"
	"github.com/gin-gonic/gin"
	"net/http"
)

func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func testConn(ctx *gincontext.Context) {
	result := base.HelloResult{Message: "connected"}

	ctx.ResSucData(result)
}
