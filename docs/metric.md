## cpu 指标

| 指标                   | 描述                                                                                                                        | 释义                        | 单位  |
|:---------------------|:--------------------------------------------------------------------------------------------------------------------------|:--------------------------|:----|
| cpu_usage_active     | Percentage of time that the CPU is active in any capacity.                                                                | CPU 活跃的时间百分比              | 百分比 |
| cpu_usage_user       | Percentage of time that the CPU is in user mode.                                                                          | CPU 处于用户态的时间百分比           | 百分比 |
| cpu_usage_system     | Percentage of time that the CPU is in system mode.                                                                        | CPU 处于系统态的时间百分比           | 百分比 |
| cpu_usage_idle       | Percentage of time that the CPU is idle.                                                                                  | CPU 空闲的时间百分比              | 百分比 |
| cpu_usage_nice       | Percentage of time that the CPU is in user mode with low-priority processes.                                              | CPU 处于低优先级进程的用户态的时间百分比    | 百分比 |
| cpu_usage_iowait     | Percentage of time that the CPU is waiting for I/O operations to complete.                                                | CPU 等待 I/O 操作完成的时间百分比     | 百分比 |
| cpu_usage_irq        | Percentage of time that the CPU is servicing interrupts.                                                                  | CPU 处理中断的时间百分比            | 百分比 |
| cpu_usage_softirq    | Percentage of time that the CPU is servicing software interrupts.                                                         | CPU 处理软件中断的时间百分比          | 百分比 |
| cpu_usage_steal      | Percentage of time that the CPU is in stolen time, or time spent in other operating systems in a virtualized environment. | CPU 在虚拟化环境中用于其它操作系统的时间百分比 | 百分比 |
| cpu_usage_guest      | Percentage of time that the CPU is running a virtual CPU for a guest operating system.                                    | CPU 用于客户操作系统的时间百分比        | 百分比 |
| cpu_usage_guest_nice | Percentage of time that the CPU is running a virtual CPU for a guest operating system with a low priority.                | CPU 用于优先级较低的客户操作系统的时间百分比  | 百分比 |

## disk 指标

| 标签     | 描述     |
|:-------|:-------|
| path   | 挂载点路径  |
| device | 设备文件   |
| fstype | 文件系统类型 |

| 指标                       | 描述                                                | 释义             | 单位  |
|:-------------------------|:--------------------------------------------------|:---------------|:----|
| disk_total_bytes         | Total number of bytes of space on the disk.       | 磁盘上的总空间字节数     | 字节  |
| disk_free_bytes          | Number of bytes of free space on the disk.        | 磁盘上可用空间的字节数    | 字节  |
| disk_used_bytes          | Number of bytes of used space on the disk.        | 磁盘上已使用空间的字节数   | 字节  |
| disk_used_percent        | Percentage of used space on the disk.             | 磁盘上已用空间的百分比    | 百分比 |
| disk_inodes_total        | Total number of index nodes reserved on the disk. | 磁盘上保留的索引节点总数   | 个   |
| disk_inodes_free         | Number of index nodes available on the disk.      | 磁盘上可用的索引节点数    | 个   |
| disk_inodes_used         | Number of index nodes used on the disk.           | 磁盘上使用的索引节点数    | 个   |
| disk_inodes_used_percent | Percentage of index nodes used on the disk.       | 磁盘上使用的索引节点的百分比 | 百分比 |

## diskio 指标

| 标签   | 描述   |
|:-----|:-----|
| name | 设备名称 |

| 指标                      | 描述                                                                                            | 释义                       | 单位 |
|:------------------------|:----------------------------------------------------------------------------------------------|:-------------------------|:---|
| diskio_reads            | Number of device read operations.                                                             | 设备读取操作的数量                | 个  |
| diskio_writes           | Number of device write operations.                                                            | 设备写入操作的数量                | 个  |
| diskio_read_bytes       | Number of bytes read from the device.                                                         | 设备读取的字节数                 | 字节 |
| diskio_write_bytes      | Number of bytes written to the device.                                                        | 设备写入的字节数                 | 字节 |
| diskio_read_time        | Number of milliseconds that read requests have waited on the device.                          | 读取请求在设备上等待的毫秒数           | 毫秒 |
| diskio_write_time       | Number of milliseconds that write requests have waited on the device.                         | 写入请求在设备上等待的毫秒数           | 毫秒 |
| diskio_io_time          | Number of milliseconds during which the device has had I/O requests queued.                   | 设备已将 I/O 请求排入队列的毫秒数      | 毫秒 |
| diskio_iops_in_progress | Number of I/O requests that have been issued to the device driver but have not yet completed. | 已向设备驱动程序发出但尚未完成的 I/O 请求数 | 个  |

## mem 指标

| 指标                    | 描述                                   | 释义        | 单位  |
|:----------------------|:-------------------------------------|:----------|:----|
| mem_total_bytes       | Total number of bytes of memory.     | 内存的总字节数   | 字节  |
| mem_available_bytes   | Number of bytes of available memory. | 可用内存的字节数  | 字节  |
| mem_used_bytes        | Number of bytes of used memory.      | 已用内存的字节数  | 字节  |
| mem_available_percent | Percentage of available memory.      | 可用内存的百分比  | 百分比 |
| mem_used_percent      | Percentage of used memory.           | 已用内存的百分比  | 百分比 |
| mem_active_bytes      | Number of bytes of active memory.    | 活跃内存的字节数  | 字节  |
| mem_buffered_bytes    | Number of bytes of buffered memory.  | 内存缓冲的字节数  | 字节  |
| mem_cached_bytes      | Number of bytes of cached memory.    | 内存缓存的字节数  | 字节  |
| mem_free_bytes        | Number of bytes of free memory.      | 空闲内存的字节数  | 字节  |
| mem_inactive_bytes    | Number of bytes of inactive memory.  | 非活跃内存的字节数 | 字节  |

## net 指标

| 标签        | 描述 |
|:----------|:---|
| interface | 网卡 |

| 指标               | 描述                                                              | 释义          | 单位 |
|:-----------------|:----------------------------------------------------------------|:------------|:---|
| net_bytes_sent   | Number of bytes sent by the network interface.                  | 网卡发送的字节数    | 字节 |
| net_bytes_recv   | Number of bytes received by the network interface.              | 网卡接收的字节数    | 字节 |
| net_packets_sent | Number of packets sent by the network interface.                | 网卡发送的数据包数   | 个  |
| net_packets_recv | Number of packets received by the network interface.            | 网卡接收的数据包数   | 个  |
| net_err_in       | Number of receive errors detected by the network interface.     | 网卡检测到的接收错误数 | 个  |
| net_err_out      | Number of transmit errors detected by the network interface.    | 网卡检测到的传输错误数 | 个  |
| net_drop_in      | Number of received packets dropped by the network interface.    | 网卡丢弃的接收数据包数 | 个  |
| net_drop_out     | Number of transmitted packets dropped by the network interface. | 网卡丢弃的传输数据包数 | 个  |
