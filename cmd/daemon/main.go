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


package main

import (
	"fmt"
	"github.com/ApsaraDB/PolarDB-Stack-Daemon/cmd/daemon/app"
	"github.com/ApsaraDB/PolarDB-Stack-Daemon/version"
	"k8s.io/component-base/logs"
	"k8s.io/klog"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func init() {
	fmt.Printf("----------------------------------------------------------------------------------------------\n")
	fmt.Printf("|                                                                                           |\n")
	fmt.Printf("| polarbox cloud branch:%v commitId:%v \n", version.GitBranch, version.GitCommitId)
	fmt.Printf("| polarbox repo %v\n", version.GitCommitRepo)
	fmt.Printf("| polarbox commitDate %v\n", version.GitCommitDate)
	fmt.Printf("|                                                                                           |\n")
	fmt.Printf("----------------------------------------------------------------------------------------------\n")
}

//程序主入口
func main() {
	rand.Seed(time.Now().UnixNano())

	command := app.NewControllerManagerCommand()

	klog.Infoln("--------------------------------------------------------------------------------------------")
	klog.Infoln("|                                                                                           |")
	klog.Infoln("|                              polarstack-daemon                                            |")
	klog.Infoln("|                                                                                           |")
	klog.Infoln("--------------------------------------------------------------------------------------------")

	logs.InitLogs()
	defer logs.FlushLogs()

	c := make(chan os.Signal)
	fmt.Println("start polarbox controller-manager cloud-provider")
	//监听指定信号
	signal.Notify(c, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGTSTP, os.Interrupt, os.Kill)

	go func() {
		//阻塞直至有信号传入
		s := <-c
		switch s {
		case syscall.SIGINT:
			klog.Infof("SIGINT")
			return
		case syscall.SIGKILL:
			klog.Infof("SIGKILL")
			return
		case syscall.SIGTERM:
			klog.Infof("SIGTERM")
			return
		case syscall.SIGSTOP:
			klog.Infof("SIGSTOP")
			return
		case syscall.SIGTSTP:
			klog.Infof("SIGTSTP")
			return
		default:
			klog.Infof("default, %v", s)
			return
		}
		fmt.Println("shut down polarstack-daemon.")
		os.Exit(1)
	}()

	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
