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

package db_log_monitor

import (
	"fmt"
	"github.com/ApsaraDB/PolarDB-Stack-Daemon/polar-controller-manager/util"
	"strings"
	"time"

	config "github.com/ApsaraDB/PolarDB-Stack-Daemon/cmd/daemon/app/config"
	utils "github.com/ApsaraDB/PolarDB-Stack-Daemon/polar-controller-manager"
	"k8s.io/kubernetes/pkg/util/slice"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
)

func StartLogMonitor(stop <-chan struct{}) {
	defer utilruntime.HandleCrash()
	klog.Infof("Starting StartLogMonitor")
	defer klog.Infof("Shutting down StartLogMonitor")
	go wait.Until(checkInsFolderTask, 6*time.Hour, stop)
	<-stop
}

func checkInsFolderTask() {
	defer utilruntime.HandleCrash()
	start := time.Now()
	defer func() {
		klog.Infof("checkInsFolderTask done, spend: %v s", time.Now().Sub(start).Seconds())
	}()

	podInsIdList := getInsIdListFromPod()
	if podInsIdList == nil {
		return
	}
	klog.Info("get insId from pods [%v]", podInsIdList)
	nodeInsIdList := getInsIdListFromNode()
	klog.Info("get insId from nodes [%v]", nodeInsIdList)
	for _, insId := range podInsIdList {
		if _, ok := nodeInsIdList[insId]; ok {
			delete(nodeInsIdList, insId)
		}
	}
	var overdueInsIdList []string
	var notOverdueInsIdList []string
	for insId, notOverdue := range nodeInsIdList {
		if !notOverdue {
			overdueInsIdList = append(overdueInsIdList, insId)
		} else {
			notOverdueInsIdList = append(notOverdueInsIdList, insId)
		}
	}
	deleteInsFolder(overdueInsIdList)
	klog.Infof("instance folder %s [%v] on %s have been deleted", config.Conf.DbclusterLogDir, overdueInsIdList, config.Conf.CurrentNodeName)
	deleteInsLogs()
	deleteRemovedInsData()
	klog.Infof("instance folder %s [%v] on %s overdue logs and data have been deleted", config.Conf.DbclusterLogDir, notOverdueInsIdList, config.Conf.CurrentNodeName)
}

func deleteInsLogs() error {
	cmd := fmt.Sprintf(`find %s -name "postgresql*.log" -mtime +%d -exec rm -f {} \;`, config.Conf.DbclusterLogDir, config.Conf.InsFolderOverdueDays)
	return utils.ExecCommand(config.Conf.CurrentNodeName, func(out string, err error) bool {
		return err == nil
	}, cmd)
}

func deleteRemovedInsData() error {
	cmd := fmt.Sprintf(`find %s -name 'rm_data_*' -type d -mtime +%d -exec rm -fr {} \;`, config.Conf.DbclusterLogDir, config.Conf.InsFolderOverdueDays)
	return utils.ExecCommand(config.Conf.CurrentNodeName, func(out string, err error) bool {
		return err == nil
	}, cmd)
}

func deleteInsFolder(insIds []string) error {
	cmd := fmt.Sprintf("cd %s; rm -fr", config.Conf.DbclusterLogDir)
	exec := false
	for _, insId := range insIds {
		if insId != "" {
			cmd += fmt.Sprintf(" %s", insId)
			exec = true
		}
	}
	if !exec {
		return nil
	}
	return utils.ExecCommand(config.Conf.CurrentNodeName, func(out string, err error) bool {
		return err == nil
	}, cmd)
}

func getInsIdListFromPod() []string {
	var allPodInsId []string
	allInsPods, err := util.GetPodsByLabels("apsara.metric.ins_id", "")
	if err != nil {
		klog.Error(err)
		return allPodInsId
	}
	for _, insPod := range allInsPods.Items {
		insId := insPod.Labels["apsara.metric.ins_id"]
		if insId != "" {
			allPodInsId = append(allPodInsId, insId)
		}
	}
	return allPodInsId
}

func getInsIdListFromNode() map[string]bool {
	insTimeList := map[string]bool{}
	var allInsResult string
	utils.ExecCommand(config.Conf.CurrentNodeName, func(out string, err error) bool {
		allInsResult = out
		return err == nil
	}, []string{
		"ls " + config.Conf.DbclusterLogDir,
	}...)
	allInsList := strings.Split(allInsResult, "\n")

	replacedDir := strings.Replace(config.Conf.DbclusterLogDir, "/", "\\/", -1)
	cmdList := []string{
		fmt.Sprintf(`find %s -name "*.log" -mtime -%d|sed 's/%s//g'|awk -F"/" '{print $1}'|uniq`,
			config.Conf.DbclusterLogDir,
			config.Conf.InsFolderOverdueDays,
			replacedDir),
	}
	var result string
	utils.ExecCommand(config.Conf.CurrentNodeName, func(out string, err error) bool {
		result = out
		return err == nil
	}, cmdList...)
	items := strings.Split(result, "\n")

	for _, item := range allInsList {
		item = strings.TrimSpace(item)
		if item != "" {
			insTimeList[item] = slice.ContainsString(items, item, nil)
		}
	}
	return insTimeList
}
