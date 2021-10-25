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

package timer

import (
	"fmt"
	"net"
	"testing"
)

func TestParseRange(test *testing.T) {
	range1Expected := rangePort{1232, 1232}
	CommonRangeCompare("1232", range1Expected, test)

	range2Expected := rangePort{9, 32}
	CommonRangeCompare("-9-32", range2Expected, test)

	range3Expected := rangePort{65535, 65536}
	CommonRangeCompare("65535-65536", range3Expected, test)
	range4Expected := rangePort{32, 32}
	CommonRangeCompare("32", range4Expected, test)

	range5Expected := rangePort{0, 0}
	CommonRangeCompare("04-abcd", range5Expected, test)

	range6Expected := rangePort{2000, 3000}
	CommonRangeCompare("02000-3000", range6Expected, test)
}

func CommonRangeCompare(rangx string, range1Expected rangePort, test *testing.T) {
	range1Actual := parseRange(rangx)
	if range1Expected == range1Actual {
		test.Log("passed")
	} else {
		test.Errorf("expected:%v, actual:%v", range1Expected, range1Actual)
	}
}

func TestIsPortAvailable(test *testing.T) {

	var res []rangePort

	res = append(res, rangePort{5400, 5800}, rangePort{15400, 15800}, rangePort{30000, 32999}, rangePort{36000, 39999}, rangePort{33000, 35999})

	//fmt.Printf("netstat -tunpa | egrep \"tcp|udp\"| awk '{if ($7==\"\") print $4\"->\"$6; else print $4\"->\"$7}' | awk '{match($0,/.+:([^,]+)->/,a);print a[1]\" \"$0}'|awk '{if ($1!=\"22\" %v) print}'|sort", buildAwkShellParts(res))

	var port = 5433
	isNotUsed := isPortAvailable(port)
	if isNotUsed {
		test.Logf("port:%d is not used, it will be listened, then test isPortAvailable method", port)
		address := fmt.Sprintf("%s:%d", "0.0.0.0", port)
		listener, err := net.Listen("tcp", address)

		if err != nil {
			test.Errorf("expected that listening on port %d is done, but actual is err:%v", port, err)
		}

		defer listener.Close()
	}

	test.Logf("port:%d is used now", port)
	var testRangePorts []rangePort
	testRangePorts = append(testRangePorts, rangePort{port, port + 1})
	alreadyUsePort := scanRangePort(testRangePorts)
	if len(alreadyUsePort) == 1 && alreadyUsePort[0] == port {
		test.Log("passed")
	} else {
		test.Errorf("expected %d is used, but actual is %v", port, alreadyUsePort)
	}

}
