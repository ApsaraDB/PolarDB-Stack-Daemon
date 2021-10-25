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

package alicloud

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"k8s.io/klog"
)

const PublicKeyFilePath = "/root/.ssh/id_rsa"

var (
	SSHUserName = "root"
)

func PublicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

func SSHConnect(user, host string, port int) (*ssh.Client, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		client       *ssh.Client
		err          error
	)

	auth = append(auth, PublicKeyFile(PublicKeyFilePath))
	hostKeyCallBack := func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		return nil
	}

	clientConfig = &ssh.ClientConfig{
		User:            user,
		Auth:            auth,
		Timeout:         30 * time.Second,
		HostKeyCallback: hostKeyCallBack,
	}

	addr = fmt.Sprintf("%s:%d", host, port)

	if client, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		klog.Errorf("failed to ssh to addr:%v, clientConfig:%v, err:%v", addr, clientConfig, err)
		return nil, err
	}

	return client, nil
}

type SSHConnection struct {
	user       string
	host       string
	port       int
	TagStr     string
	client     *ssh.Client
	createTime time.Time
	Counter    int
}

func NewSSHConnection(user, host string, port int, tags ...string) *SSHConnection {
	tagStr := ""
	if len(tags) > 0 {
		tagStr = "[" + strings.Join(tags, "|") + "]"
	}
	return initSSHClient(user, host, port, tagStr)
}

func NewSSHConnectionByHost(host string, tags ...string) *SSHConnection {
	return NewSSHConnection(SSHUserName, host, 22, tags...)
}

func initSSHClient(_user, _host string, _port int, _tagStr string) (sshClient *SSHConnection) {
	return &SSHConnection{
		user:   _user,
		host:   _host,
		port:   _port,
		TagStr: _tagStr,
	}
}

func (conn *SSHConnection) Init() error {
	if conn.client != nil {
		conn.Close()
	}
	start := time.Now()
	klog.Infof("%s begin build %s ssh connection...", conn.TagStr, conn.host)
	client, err := SSHConnect(conn.user, conn.host, conn.port)
	if err != nil {
		klog.Errorf("%s begin build %s ssh connection err: %v", conn.TagStr, conn.host, err)
		return err
	}
	klog.Infof("%s build %s ssh connection done spend [%v s]", conn.TagStr, conn.host, time.Now().Sub(start).Seconds())
	conn.createTime = time.Now()
	conn.client = client
	conn.Counter = 0
	return nil
}

func (conn *SSHConnection) IsInit() bool {
	return conn.client != nil
}

func (conn *SSHConnection) TestAlive() bool {
	if !conn.IsInit() {
		return false
	}

	session, err := conn.client.NewSession()
	if err != nil {
		return false
	}

	defer func() {
		if session != nil {
			cErr := session.Close()
			if cErr != nil {
				//
			}
		}
	}()

	return true
}

func (conn *SSHConnection) ResetTime() {
	conn.createTime = time.Now()
	conn.Counter = 0
}

func (conn *SSHConnection) Close() {
	klog.V(5).Infof("%s close %s ssh connection...", conn.TagStr, conn.host)

	if conn.client != nil {
		err := conn.client.Close()
		if err != nil {
			//
		}
		conn.client = nil
	}
}

func (conn *SSHConnection) RunCmdWithLogLevel(cmd string, autoClose bool, logLevel int) (string, string, error) {

	if autoClose {
		defer conn.Close()
	}

	if !conn.IsInit() {
		klog.Errorf("[%s] please init ssh connection on[%s] first!", conn.TagStr, conn.host)
		return "", "", errors.New(fmt.Sprintf("%s please init ssh connection first", conn.TagStr))
	}

	start := time.Now()

	var stdOut, stdErr bytes.Buffer

	session, err := conn.client.NewSession()
	if err != nil {
		// close client connection if new session has failed
		klog.Errorf("%s ssh create new session on [%s] err :%v", conn.TagStr, conn.host, err)
		return "", "", err
	}

	defer func() {
		if session != nil {
			cErr := session.Close()
			if cErr != nil {
				//
			}
		}
	}()

	session.Stdout = &stdOut
	session.Stderr = &stdErr

	conn.Counter = conn.Counter + 1

	err = session.Run(cmd)
	if err != nil {
		klog.V(5).Infof("%s exec cmd[%s]:[%s] err: %v", conn.TagStr, cmd, conn.host, err)
	}

	stdOutStr := stdOut.String()
	stdErrStr := CutOffWarningLine(stdErr.String())

	stdOutStrWithNoEnter := strings.ReplaceAll(stdOutStr, "\n", "\\n")
	stdErrStrWithNoEnter := strings.ReplaceAll(stdErrStr, "\n", "\\n")

	now := time.Now()

	var totalSpend float64 = 0
	if &conn.createTime != nil {
		totalSpend = now.Sub(conn.createTime).Seconds()
	}

	spendTime := now.Sub(start).Seconds()

	_logLevel := logLevel

	if totalSpend >= 2 || spendTime >= 2 {
		//超过2秒的，要升一级提示
		_logLevel = _logLevel - 1
		if logLevel <= 1 {
			_logLevel = 1
		}
	}

	klog.V(klog.Level(_logLevel)).Infof("%s runSsh:%d cost[%v s],total[%v s] command [%s]:[%s], out:[[%s]], errOut:[[%s]], err:[%v]", conn.TagStr, conn.Counter, spendTime, totalSpend, cmd, conn.host, stdOutStrWithNoEnter, stdErrStrWithNoEnter, err)

	return stdOutStr, stdErrStr, err
}

// RunSSHNoPwdCMD run shell command
func RunSSHNoPwdCMD(cmd string, remoteAddress string, tags ...string) (string, string, error) {
	tagStr := ""
	if len(tags) > 0 {
		tagStr = "[" + strings.Join(tags, "|") + "]"
	}

	controllerConf, err := GetControllerConfig()
	if err != nil {
		klog.Errorf("%s runSsh command [%s] failed, get ssh user and password error:[%v]", tagStr, cmd, err)
		return "", "", err
	}

	connBeginTime := time.Now()
	var stdOut, stdErr bytes.Buffer
	klog.Infof("%s runSsh command [%s], on [%s]", tagStr, cmd, remoteAddress)

	client, err := SSHConnect(controllerConf.SshUser, remoteAddress, 22)

	if err != nil {
		klog.Errorf("%s runSsh command [%s] failed on build %s connect:[%v]", tagStr, cmd, remoteAddress, err)
		return "", "", err
	}

	defer func() {
		if client != nil {
			cErr := client.Close()
			if cErr != nil {
				//
			}
		}
	}()

	session, err := client.NewSession()

	if err != nil {
		// close client connection if new session has failed
		klog.Errorf("%s cmd=[%s] create session on %s err:%v", tagStr, cmd, remoteAddress, err)
		return "", "", nil
	}

	defer func() {
		if session != nil {
			session.Close()
		}
	}()

	cmdStartTime := time.Now()

	session.Stdout = &stdOut
	session.Stderr = &stdErr

	runErr := session.Run(cmd)
	if runErr != nil {
		klog.Errorf("[%s]exec cmd[%s]:[%s] %v", tagStr, cmd, remoteAddress, runErr)
	}

	stdOutStr := stdOut.String()
	stdErrStr := CutOffWarningLine(stdErr.String())

	stdOutStrWithNoEnter := strings.ReplaceAll(stdOutStr, "\n", "\\n")
	stdErrStrWithNoEnter := strings.ReplaceAll(stdErrStr, "\n", "\\n")

	now := time.Now()

	subString := ShowSubString(stdOutStrWithNoEnter, 40000)

	klog.Infof("%s runSsh cost[%v s],total[%v s] command [%s]:[%s], out:[[%s]] errOut:[[%s]] err:[%v]", tagStr, now.Sub(cmdStartTime).Seconds(), now.Sub(connBeginTime).Seconds(), cmd, remoteAddress, subString, stdErrStrWithNoEnter, err)

	return stdOutStr, stdErrStr, nil
}

func ShowSubString(orgText string, l int) string {
	if l <= 0 {
		return orgText
	}
	if len(orgText) <= l {
		return orgText
	}

	bts := []byte(orgText)

	return string(bts[:l]) + "..."
}

func CutOffWarningLine(orgErrStr string) string {
	const warnHead = "Warning: Permanently added"
	const warnEnd = "to the list of known hosts."

	pos := strings.Index(orgErrStr, warnEnd)
	if pos < 0 {
		return orgErrStr
	}

	head := orgErrStr[:pos]

	if !strings.Contains(head, warnHead) {
		//如未包括头字符，暂不确定结果是否正确
		return orgErrStr
	}

	r := strings.TrimSpace(orgErrStr[pos+len(warnEnd):])

	if &r == nil {
		r = ""
	}

	return r

}

func ExecCommand(runCmdNode string, checkSucceedFunc func(string, error) bool, cmdList ...string) error {
	for _, cmd := range cmdList {
		var err error
		sshOut, errMsg, sshErr := RunSSHNoPwdCMD(cmd, runCmdNode)
		out := sshOut
		if sshErr != nil {
			err = sshErr
		} else if errMsg != "" {
			err = errors.New(errMsg)
		}

		if !checkSucceedFunc(out, err) {
			errorInfo := fmt.Sprintf("execute cmd [%s] result: [%s], [%v]", cmd, out, err)
			klog.Error(errorInfo)
			return err
		}
	}
	return nil
}
