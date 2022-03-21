## What is PolarDB Stack Daemon?

PolarDB Stack Daemon is a component in PolarStack, an Alibaba Cloud's DBaaS hybrid cloud product. It runs on each host and is mainly responsible for the following operations:

1. Port scanning: PolarDB Stack Daemon regularly scans the port occupation status of the host where it is located and provides basic data for the database engine and other components to allocate ports.
2. Kernel image version scanning: PolarDB Stack Daemon identifies the availability of the kernel version on the machine where it is located and provides basic data for database recovery and specifying a specific version when creating a database.
2. Database log cleaning: For the database sharing storage among the primary node and read-only nodes based on the Shared-Storage architecture, the data is stored on the shared storage and the database logs are stored on the local disk of the host. Cleaning logs on the local disk regularly according to the rules is necessary for ensuring the long-term stable operation of the host and the database.
4. Node network status and host status collecting: PolarDB Stack Daemon queries the status of the host client network at the regular time, updates the specified node condition, and provides the IP information and network status information required for the connection to the database engine and the database proxy consisting of a primary node and at least one read-only nodes.

## Branch Introduction

The `main` is the default branch of PolarDB Stack Daemon.

## Code Structure

PolarDB Stack Daemon project uses cobra.Command to start the program, which mainly consists of scheduled tasks and HTTP services. The startup part of the project code is in the cmd directory and the business code is in the polar-controller-manager directory.

The polar-controller-manager directory mainly consists of the following subdirectories:

- port_usage: mainly executes port scanning and summarizes the port occupation status of the host into the configmap.
- core_version: scans the image information of the minor version of the kernel on the local machine when the program starts. Then the image of the minor version of the kernel will be scanned again according to the API calling and summarized into the specific configmap.
- db_log_monitor: contains the code for cleaning database engine logs.
- node_net_status: mainly updates the status of Kubernetes node condition regularly, including NodeClientIP, NodeClientNetworkUnavailable, and NodeRefreshFlag. The NodeClientIP condition stores the IP information of the client network for the database to directly communicate with cm or the proxy.
- bizapis: provides RESTful services. The specific implementation is in the service directory under this directory. The corresponding implementation functions in other directories can be called by functions in the service directory.

## Quick Start

We provide two methods to use PolarDB:

- Alibaba Cloud PolarDB (Hybrid Cloud Edition)
- Deploy open-source PolarStack locally

### Alibaba Cloud PolarDB (Hybrid Cloud Edition)

Alibaba Cloud PolarDB (Hybrid Cloud Edition): [Official Website](https://www.alibabacloud.com/product/polarbox).

### Deploy an Instance Running Locally

**Before You Start**

PolarDB Stack Daemon runs on each node machine in the form of Kubernetes daemonset and executes operation commands on the machine via SSH. Before deployment, make sure that Kubernetes is installed, all components of Kubernetes are running normally, and all hosts have set up SSH passwordless access to each other.

**Steps**

1. Download the source code of PolarDB Stack Daemon from <https://github.com/ApsaraDB/PolarDB-Stack-Daemon>.

2. Install Kubernetes and ensure that all Kubernetes components are running normally.

3. Install related dependencies by executing the command `kubectl create -f all.yaml`.

   *Note:* This YAML file contains all contents required for PolarDB Stack Daemon deployment.
   
   a. NIC configuration ccm-config configmap:
   
   - Set NET_CARD_NAME to the name of the business network NIC and NET_MASK to the subnet mask of the business network NIC.
   
   b. Create ClusterRole, ServiceAccount, and ClusterRoleBinding required for PolarDB Stack Daemon running.
   
   - ClusterRole: cloud-controller-manager
   - ClusterRoleBinding: cloud-controller-manager
   
   - ServiceAccount: cloud-controller-manage

   c. Main startup parameters:
   
   - Parameters of the directory where database logs are stored: dbcluster-log-dir.
   
   - Rules of cleaning database logs (unit: day): ins-folder-overdue-days.
   
   - Labels of configmap where the minor version information of the kernel is stored: core-version-cm-labels.
   

   d. Kubernetes daemonset settings:
   
   - Access some commands of the host via SSH. The /root/.ssh is mounted.
   
   - The directory where the polarstack-daemon logs are stored is /var/log/polardb-box/polardb-net, which is mounted.
   
   - Access the local network because the daemon-set uses the host network (hostNetwork: true).
   
4. Check the running status.

​		After deploying PolarDB Stack Daemon, you can check the daemonset pod status, the condition status in the Kubernetes node, port scanning and cleaning configmap in Kubernetes, and the existence configmap of the kernel image version to check whether it functions well.

​		a. Check the pod running status of PolarDB Stack Daemon. There is a pod of polarstack-daemon on each host and you need to check whether they are all in the running status.

​    ```kubectl get pod -owide -A |grep polarstack-daemon```

![img](docs/img/1.png)

​		b. Check the port scanning status. Each polarstack-daemon pod will identify the port occupation status of the host by trying to listen on the ports and store the used ports in the configmap.

​    ```kubectl get cm -A |grep port-usage```

![img](docs/img/2.png)

​		c. Check the kernel version. PolarDB Stack Daemon queries the configmap of the minor version information of the kernel according to the parameter value during startup and then queries whether the image information exists on the host according to the configmap.

​     ```kubectl get cm -A |grep version-availability```

![img](docs/img/3.png)

As shown in the figure below, two minor versions 11.2.20200630.0172e3f3.20201103225317 and 11.2.20200630.e0eb5bdb.20210317155810 of the kernel exist on the host polardb-box-soft011160139051.

![img](docs/img/4.png)

## Contributions

You are welcome to make contributions to PolarDB Stack Daemon. We appreciate all the contributions. For more information about how to start development and pull requests, see [contributing](https://github.com/ApsaraDB/PolarDB-for-PostgreSQL/blob/main/doc/PolarDB-EN/contributing.md).

## Software License

PolarDB Stack Daemon's code is released under the Apache License (Version 2.0). See the [LICENSE](https://github.com/ApsaraDB/PolarDB-for-PostgreSQL/blob/main/LICENSE.txt) and [NOTICE](https://github.com/ApsaraDB/PolarDB-for-PostgreSQL/blob/main/NOTICE.txt) for more information.

## Acknowledgments

Some codes and design ideas are based on other open-source projects, such as Kubernetes and Gin. We thank the contributions of the preceding open-source projects.

## Contact us

Use the DingTalk application to scan the following QR code and join the DingTalk group for PolarDB technology promotion.

![QR Code](docs/img/QRCode.png)

