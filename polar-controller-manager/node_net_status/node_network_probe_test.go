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


package node_net_status

import (
	"fmt"
	"testing"
)

func TestFormatSanOutput(t *testing.T) {
	out1 := "Warning: Permanently added '198.18.3.6' (ECDSA) to the list of known hosts.\n" +
		"h_r03.dbm-01|online\n" +
		"h_r03.dbm-02|offline\n" +
		"h_r03.dbm-03|xxxx\n"
	hostName1 := "r03.dbm-01"
	hostName2 := "r03.dbm-02"
	hostName3 := "r03.dbm-03"
	hostName4 := "r03.dbm-04"

	mapStorageStatus := _FormatSanOutput(out1)
	if mapStorageStatus == nil {
		t.Error("no storageStatus info")
	}

	storageStatus1, ok := (*mapStorageStatus)[hostName1]
	if ok {
		if storageStatus1.Available {
			t.Log("passed")
		} else {
			t.Errorf("hostName:%s should be in 'true' state, but it is status:%v", hostName1, *storageStatus1)
		}
	} else {
		t.Errorf("hostName:%s should exist", hostName1)
	}

	storageStatus2, ok := (*mapStorageStatus)[hostName2]
	if ok {
		if !storageStatus2.Available {
			t.Log("passed")
		} else {
			t.Errorf("hostName:%s should be in 'false' state, but it is status:%v", hostName2, storageStatus2)
		}
	} else {
		t.Errorf("hostName:%s should exist", hostName2)
	}

	storageStatus3, ok := (*mapStorageStatus)[hostName3]
	if ok {
		if !storageStatus3.Available {
			t.Log("passed")
		} else {
			t.Errorf("hostName:%s should be in 'false' state, but it is status:%v", hostName3, storageStatus3)
		}
	} else {
		t.Errorf("hostName:%s should exist", hostName3)
	}

	StorageStatus4, ok := (*mapStorageStatus)[hostName4]
	if ok {
		t.Errorf("hostName:%s should not exist, but it is status:%v", hostName4, StorageStatus4)
	} else if !ok {
		t.Log("passed")
	}

	fmt.Printf("%v\n", mapStorageStatus)
}
