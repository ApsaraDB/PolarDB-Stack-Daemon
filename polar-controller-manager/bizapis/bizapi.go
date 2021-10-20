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
	"fmt"
	"github.com/gin-gonic/gin"
	config "github.com/ApsaraDB/PolarDB-Stack-Daemon/cmd/daemon/app/config"
	"github.com/ApsaraDB/PolarDB-Stack-Daemon/polar-controller-manager/bizapis/controller"
	"github.com/ApsaraDB/PolarDB-Stack-Daemon/polar-controller-manager/bizapis/service"
	"github.com/ApsaraDB/PolarDB-Stack-Daemon/polar-controller-manager/core_version"
	"github.com/ApsaraDB/PolarDB-Stack-Daemon/polar-controller-manager/util"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

const (
	PathHealthz                 = "/healthz"
	PathInnerCheckCoreVersion   = "InnerCheckCoreVersion"
	PathRequestCheckCoreVersion = "RequestCheckCoreVersion"
	PathGetStandByIp            = "GetStandByIp"
	PathTestConn                = "TestConn"
)

func StartHttpServer(cfg *config.CompletedConfig, client kubernetes.Interface) {
	port := cfg.Port
	klog.Infof("webserver port:%d", port)
	router := gin.Default()
	componentService := service.NewService(client, cfg)
	initRoute(router, componentService)

	go func() {
		gin.SetMode(gin.ReleaseMode)

		addr := fmt.Sprintf(":%d", port)
		err := router.Run(addr)
		if err != nil {
			klog.Warningf("failed to start webserver, err:%v", err)
		}
	}()

	klog.Infof("start webserver on port:%d done", port)
}

func initRoute(router *gin.Engine, service *service.Service) {
	router.GET(PathHealthz, health)

	v1Group := router.Group(util.PathPrefix)

	systemCtl := controller.NewSystemController(service.System)

	GET(v1Group, PathTestConn, testConn, PublicAPI, "test conn")
	GET(v1Group, PathGetStandByIp, systemCtl.GetPodStandByIp, PublicAPI, "get standby ip")
	POST(v1Group, PathRequestCheckCoreVersion, core_version.RequestCheckCoreVersion, PublicAPI, "request to check core version")
	POST(v1Group, PathInnerCheckCoreVersion, core_version.InnerCheckCoreVersion, PublicAPI, "inner request to check core version")
}
