## 什么是PolarDB Stack Daemon？

PolarDB Stack Daemon（下文简称为ps-daemon）是阿里云DBaaS混合云产品PolarStack中负责主机上的组件，会在每台主机上运行，主要负责以下操作：

1. 端口扫描：定期扫描所在主机的端口占用情况，为数据库引擎以及其他组件分配端口提供基础数据。
2. 内核镜像版本扫描：识别所在机器上内核版本的可用性，为数据库恢复和创建数据库时指定特定版本提供基础数据。
3. 数据库日志清理：基于共享存储的一写多读数据库数据在共享存储上，数据库的日志主机的本地盘上，定期按照标准清理保持本地盘的日志是保证主机与数据库长期稳定运行的重要手段。
4. 节点网络情况与主机情况收集：定时查询所在主机客户网网络情况，更新指定node condition，为数据库引擎、一些多读数据库代理提供连接所需的IP信息与网络状态信息。

## 分支说明

PolarDB Stack Daemon后续默认分支为main分支。

## 代码结构

PolarDB Stack Daemon工程采用cobra.Command形式启动程序， 主要由一些定时任务和http服务组成， 工程代码结构启动部分在目录cmd， 业务代码在polar-controller-manager目录。

polar-controller-manager目录主要由以下子目录组成：

- port_usage目录：主要执行端口扫描，将本机上端口已占用情况汇总到configmap中；
- core_version目录：启动扫描一次本机上内核小版本的image信息，之后根据调用再次执行内核小版本image扫描，并汇总到执行configmap；
- db_log_monitor目录： 数据库引擎日志清理的代码；
- node_net_status目录： 主要定时更新k8s node condition状态，包括NodeClientIP、NodeClientNetworkUnavailable、 NodeRefreshFlag， 其中NodeClientIP condition会存储客户网IP信息，供给数据库与cm或者代理直接通信使用；
- bizapis目录：对外提供的restful服务，可在该目录下service中可调用其他目录对应内容；

## 快速开始

我们提供了两种途径来使用PolarDB数据库：

- 阿里云PolarDB混合云版本。
- 本地搭建开源PolarStack。

### 阿里云PolarDB混合云版本

阿里云PolarDB 混合云版：[官网地址](https://www.alibabacloud.com/zh/product/polarbox)。

### 搭建本地运行的实例

**操作前提：**

PolarDB Stack Daemon以k8s daemonset形式运行在每台node机器上，通过ssh在本机上执行操作命令。部署前，请确保k8s安装完毕，k8s各组件运行正常， 且所有主机之间互相已打通免密访问。

**步骤：**

1. 下载PolarDB Stack Daemon源代码，地址：http://xxxx。
2.  安装k8s，确保k8s个组件处于正常运行状态。
3.  安装相关依赖：执行kubectl create -f all.yaml

*说明：*该yaml文件包含了PolarStack-Daemon部署所需的全部内容。

a, 网卡配置ccm-config configmap：

- - 配置了NET_CARD_NAME业务网网卡名称。NET_MASK业务网网卡子网掩码

b, 创建了PolarDB Stack Daemon运行所需的ClusterRole、ServiceAccount、ClusterRoleBinding

- - ClusterRole：cloud-controller-manager
- ClusterRoleBinding：cloud-controller-manager

- - ServiceAccount：cloud-controller-manage

c, 主要启动参数：

- - 数据库日志所在目录参数dbcluster-log-dir
- 数据库日志清理标准（单位天）ins-folder-overdue-days

- - 内核小版本信息所在configmap的label标签：core-version-cm-labels

d, k8s daemonset设置：

- - 需要通过ssh访问本机的一些命令，挂载了/root/.ssh
- polarstack-daemon日志所在目录/var/log/polardb-box/polardb-net， 挂载了该目录

- - 需要访问本机网情况，因为daemon-set使用了主机网络hostNetwork: true

1. 检查运行状态

​     部署完PolarDB Stack Daemon后，可以通过查看daemonset pod状态，k8s node中的condition状态，k8s中端口扫描清理configmap， 内核镜像版本存在性configmap查看功能是否正常

​    a, PolarDB Stack Daemon的pod运行情况, 每台机器上有一个polarstack-daemon的pod， 都处于runningzhaungtai

​    kubectl get pod -owide -A |grep polarstack-daemon

![img](https://intranetproxy.alipay.com/skylark/lark/0/2021/png/288373/1632647922395-52bb9f96-9f03-444b-9e43-5b63d60c6782.png)

​    b, 查看端口扫描情况， 每个polarstack-daemon pod会通过尝试监听端口的方式识别本机上端口占用情况，并将已使用端口存入configmap中

​    kubectl get cm -A |grep port-usage

![img](https://intranetproxy.alipay.com/skylark/lark/0/2021/png/288373/1632648185711-686a7a48-9fc6-4814-beab-57ab2032e359.png)

​    c, 查看内核版本情况，PolarDB Stack Daemon在启动时会根据参数值查询内核小版本信息的configmap，然后根据configmap查询本机上是否存在这些image信息

​     kubectl get cm -A |grep version-availability

![img](https://intranetproxy.alipay.com/skylark/lark/0/2021/png/288373/1632648411948-c7a5a08b-d385-42fb-a984-50f2fa038aef.png)

如下所示图中表示两个内核小版本11.2.20200630.0172e3f3.20201103225317和11.2.20200630.e0eb5bdb.20210317155810存在于polardb-box-soft011160139051机器上



![img](https://intranetproxy.alipay.com/skylark/lark/0/2021/png/288373/1632648640395-7c4e5fb8-ff89-4cfe-b46c-b0be6fb6cac4.png)

## 贡献

我们非常欢迎和感激您的贡献，请参见

[contributing]:

一文来了解如何开始开发以及pull request。

## 软件许可说明

PolarDB Stack Daemon的代码的发布基于Apache 2.0版本软件许可。相关的许可说明可参见[License](https://github.com/alibaba/PolarDB-for-PostgreSQL/blob/master/LICENSE)和[NOTICE](https://github.com/alibaba/PolarDB-for-PostgreSQL/blob/master/NOTICE)。

## 致谢

部分代码和设计思路参考了其他开源项目，例如：kubernets、Gin。感谢以上开源项目的贡献。

## 联系我们

使用钉钉扫描如下二维码，加入PolarDB技术推广组钉钉群。

