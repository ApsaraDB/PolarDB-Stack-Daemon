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
	"gitlab.alibaba-inc.com/rds/polarstack-daemon/polar-controller-manager/core_version"
	"gitlab.alibaba-inc.com/rds/polarstack-daemon/polar-controller-manager/util"
	"io/ioutil"
	"net/http"
	"testing"
)

func Test_CheckCoreVersion(t *testing.T) {
	router := gin.New()
	v1Group := router.Group(util.PathPrefix)

	POST(v1Group, PathRequestCheckCoreVersion, core_version.RequestCheckCoreVersion, PublicAPI, "request to check core version")
	POST(v1Group, PathInnerCheckCoreVersion, core_version.InnerCheckCoreVersion, PublicAPI, "inner request to check core version")

	var port = 8900
	addr := fmt.Sprintf(":%d", port)
	go router.Run(addr)
	CommonHttpRequestTest(port, PathRequestCheckCoreVersion, t)
	CommonHttpRequestTest(port, PathInnerCheckCoreVersion, t)
}

func CommonHttpRequestTest(port int, pathSuffix string, t *testing.T) {
	fullPath := fmt.Sprintf("http://127.0.0.1:%d%s/%s", port, util.PathPrefix, pathSuffix)
	httpReq, err := http.NewRequest(http.MethodPost, fullPath, nil)
	if err != nil {
		t.Errorf("failed to make req, path:%s, err:%v", fullPath, err)
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		t.Errorf("failed to do http req, path:%s, err:%v", fullPath, err)
	}
	if resp != nil {
		defer resp.Body.Close()
	}

	body, err := ioutil.ReadAll(resp.Body)
	if http.StatusOK != resp.StatusCode {
		if err != nil {
			t.Errorf("httpStatusCode is not %d, path:%s, httpStatusCode:%v, readBodyErr:%v",
				http.StatusOK, fullPath, resp.StatusCode, err)
		} else {
			t.Errorf("httpStatusCode is not %d, path:%s, httpStatusCode:%v, body:%s",
				http.StatusOK, fullPath, resp.StatusCode, string(body))
		}
	} else {
		if err != nil {
			t.Logf("fullPath:%s, done, passed,readBodyErr:%v", fullPath, err)
		} else {
			t.Logf("fullPath:%s, done, passed, body:%s", fullPath, string(body))
		}

	}
}
