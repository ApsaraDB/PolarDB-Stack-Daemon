# Roadmap for Open-Source PolarDB Stack Daemon

Alibaba Cloud will continuously release functions of PolarDB Stack Daemon which are valuable to users. At present, Alibaba Cloud plans the following three versions for PolarDB Stack Daemon:

### PolarDB Stack Daemon Version 1.0

- Collect the port occupation status.

- Collect the minor version information of the kernel.

- Clean database engine logs.

- Collect the network information of the host client network.

### PolarDB Stack Daemon Version 2.0

Based on version 1.0, we continuously optimize the existing functions and provide new functions in version 2.0. For example,

- Floating IP address: For a database cluster consisting of a primary node and at least one read-only node, PolarDB Stack Daemon provides a fixed IP address for the primary (read-write) node. This IP address will be migrated along with the primary node to ensure that the connection string (it is used for the connection between the application and the database) can remain unchanged after the actual database node migrates. As well, this function can provide database services for old applications with only one IP address.
- Event uploading: PolarDB Stack Daemon can upload the operations and results to the event center when performing important or key steps. These events will help the administrator learn about the system running status and troubleshoot problems.

### PolarDB Stack Daemon Version 3.0

In version 3.0, we plan to provide VIP functions. For example,

- Keepalived-based VIP functions: PolarDB Stack Daemon will provide VIP functions offline based on Keepalived and LVS.
- Existence compensation of kernel images: For newly added physical hosts, PolarDB Stack Daemon will actively replicate their kernel images from other nodes to ensure that each host has the corresponding image during database node migration.