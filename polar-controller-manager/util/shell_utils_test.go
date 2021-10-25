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

package util

import (
	"fmt"

	"testing"
)

func TestGetShellExitCode(test *testing.T) {
	var remoteShellHead = "Process exited with status"
	var localShellHead = "exit status"
	var remoteShellHeadTrim = "  Process exited with status"
	var expectedEmpty = ""

	var remoteShellHead01 = "Process exited with status 01"
	var remoteShellHeadTrim01 = "  Process exited with status 01"
	var expected01 = "01"

	CommonTest(remoteShellHead, expectedEmpty, test)
	CommonTest(localShellHead, expectedEmpty, test)
	CommonTest(remoteShellHeadTrim, expectedEmpty, test)

	CommonTest(remoteShellHead01, expected01, test)
	CommonTest(remoteShellHeadTrim01, expected01, test)
}

func CommonTest(shellOutputString, expected string, test *testing.T) {
	res := GetShellExitCode(shellOutputString)
	if expected == res {
		test.Log("passed")
	} else {
		test.Errorf("expected:%v, actual:%v", expected, res)
		fmt.Printf("max is : %v\n", res)
	}
}
