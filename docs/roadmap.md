# PolarDB Stack Daemon开源 Roadmap

PolarDB Stack Daemon将持续发布对用户有价值的功能。当前我们计划了3个阶段：

### PolarDB Stack Daemon 1.0版本

- 端口已占用情况收集
- 内核小版本信息收集

- 数据库引擎日志清理
- 主机客户网网络信息收集

### PolarDB Stack Daemon 2.0版本

在1.0的基础上，2.0版本继续优化已有功能与提供新功能，例如：

- 浮动ip支持，针对一写多读数据库集群，针对rw节点提供一个固定ip， 该ip会随数据rw节点的迁移而迁移，从而保证应用程序连接数据库连接串无需因为数据库实际节点的迁移而更改，同时也为只能填写一个ip的老旧应用提供数据库服务。
- 事件上报，针对PolarDB Stack Daemon执行的重要或者关键步骤，向事件中心上报对应操作和结果，方便管理员了解系统运行状态，排查问题。

### PolarDB Stack Daemon 3.0版本

3.0版本主要计划提供vip功能，例如：

- 基于keepalived：在线下环境中基于keepalived和lvs提供vip功能。
- 内核镜像存在性补偿：针对新添加物理主机，主动从其他节点复制内核image保证数据库节点迁移时每台机器都有对应镜像。

