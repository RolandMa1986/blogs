---
title: Kubernetes 运维必知的内核参数和优化项
date: 2022-01-17T14:15:42Z
---

在 Kubernetes 安装过程中，除必要的内核参数外，Kubernetes 并未提供更多的优化建议。但无论 CentOS 或 Ubuntu 默认内核参数通常并不能满足高负载环境下的性能要求。因此，我们需要调整内核参数以优化节点的性能，并使节点更加稳定与健壮。
<!--more-->
## 内核&内存&文件系统

### PID Max

- `kernel.pid_max = 32768` 配置决定了系统中可用的 PID 数量。进程 ID (PID) 是 Linux 内核的一项基础资源，内核会为每个进程或线程分配一个 PID。`pid_max` 的默认值一般是由内核编译项或 CPU 数量决定的，对于 64位系统内核代码中给出的计算方法如下：
```c
# https://github.com/torvalds/linux/blob/5147da902e0dd162c6254a61e4c57f21b60a9b1c/kernel/pid.c#L657
pid_max = min(4194304, max(32768, 1024 * (number of CPUs)))
```
以上表达式可以简单理解为，如果小于32核心则默认最大PID值为32768，超过32核心的则按1024* CPUs计算。因此，1024 也是内核推荐的每核心所调度的进程数量。

在一个 Kebenetes 的工作节点即使调度了110个pod， 那么每个 Pod 也可以获得 200 多个 PID。这个限制对大多数应用通常是足够的。但 Kuberentes 集群中通常会存在资源超卖的情况，此时可以适当提高此设置。

但在节点上应用 `pid_max` 设置时，并不能防范节点上 PID 资源耗尽问题。如果某 Pod 无限制的创建进程可能会导致节点运行不稳定。因此我们还需要启用 Kubernetes 中的另一个特性限制/预留 POD 的 PID 资源。Kubelet 中提供了 `--system-reserved`, `--kube-reserved`等参数，用以设置系统和 kubelet 所需要保留的 PID 数量用来保证节点上的程序可以正常运行调度。`--pod-max-pids` 用以限制 POD 最大可分配的 PID 数量。

具体配置请参考：
> https://kubernetes.io/docs/concepts/policy/pid-limiting/

### Softlockup panic
- `kernel.softlockup_panic = 1`
- `kernel.softlockup_all_cpu_backtrace = 1`

Softlockup 通常是由 bug 引起的内核异常，导致在内核空间正在执行的任务无法调度。同时，这个任务也会阻止其他任务的执行。当发生无法恢复的异常时，重启往往是最后的恢复手段。

> https://github.com/kubernetes/kubernetes/pull/38001

### 文件限制

- `fs.file-max = 2097152` 限制了 linux 全局范围可以打开文件的个数。当 POD 中大量运行数据库或 web 服务器，导致默认的 file-max 无法满足需求时，可以修改此参数增大最大打开文件的数量。

除此，`inotify` 也是文件相关的重要资源，如 `Nginx`，`kubectl logs -f` 等都会产生并消耗 `inotify` 资源，可适当优化这些参数。

- `fs.inotify.max_user_instances = 8192`
- `fs.inotify.max_user_watches = 524288`

> https://ieevee.com/tech/2019/09/29/ulimit.html

### 内存
- `vm.max_map_count = 262144` 定义了一个进程能拥有的最多的内存区域。Elastic Search, Mongodb 等程序都使用 `mmap` 技术将文件映射到内存地址空间，用于优化文件访问。Elastic Search 推荐的 `max_map_count = 262144`，如果按内存页默认大小 4K计算，该设置可为 `mmap` 最大分配1G内存地址空间。

> https://github.com/kubernetes/kops/issues/1340
> https://www.elastic.co/guide/en/elasticsearch/reference/current/vm-max-map-count.html


### 其他默认设置

除此以外，Kubelet 会默认配置以下4个设置，用于定义应用的 oom行为和系统行为：

- `vm.overcommit_memory = 1`
- `vm.panic_on_oom = 0`

Virtual Memory 技术的应用打破了物理内存限制。基于VM技术的内存申请分为两个步骤：1. 申请内存地址空间；2.初始化物理内存。当`overcommit_memory`设置为 1 时，内核总是为应用分配内存空间地址，而不会关心物理内存时候足够。当应用需要使用物理内存，且内存不足时，由 oom_killer 进行内存回收操作，杀死超出内存限制或低优先级的进程。

- `kernel.panic = 10`
- `kernel.panic_on_oops = 1`

oops = panic 有助于系统的快速恢复，防止进一步引发未知的错误。因此，Kubernetes 中默认采用该策略。

> https://github.com/kubernetes/kubernetes/issues/12818

## 网络

### 网络优化设置
网络相关优化参数比较复杂，但总结起来可以归纳为三大类：

#### 1. 缓存配额
- `net.core.rmem_max = 16777216`
- `net.core.wmem_max = 16777216` 每个套接字的发送和接收缓存大小，此设置适用于TCP和UDP连接。
- `net.ipv4.tcp_wmem = 4096 87380 16777216`
- `net.ipv4.tcp_rmem = 4096 87380 16777216` 针对TCP连接的缓存配额，三个数值分别是最小值，默认值与最大值。最大值应小于前面全局配置。

由于系统缓存不足而导致的丢包是网络的常见问题，丢包后会造成包重传从而浪费网络带宽。我们需要针对硬件规格网络带宽进一步优化这些选项。

#### 2. 队列限制
- `net.core.somaxconn = 32768` 定义了系统中每一个端口最大的监听队列的长度, 
- `net.ipv4.tcp_max_syn_backlog = 8096` 定义了半连接队列的最大长度
- `net.core.netdev_max_backlog = 16384` 表示当每个网络接口接受数据包的速率比内核处理这些包的速率快时，允许发送队列的数据包的最大数目

当应用层来不能及时的将建立成功的TCP连接从队列中取出时,缓冲队列就会丢弃这个连接。

> https://mp.weixin.qq.com/s/Bjjrk85MTTSiHoECNIWEaw

#### 3. 其他特性
- `net.ipv4.tcp_slow_start_after_idle = 0` 禁止 TCP 拥塞窗口重新初始化，可优化长连接下的网络性能。尤其是在内网环境比较稳定，网络无拥塞的环境中。
- `net.ipv4.tcp_tw_reuse = 1` 客户端复用 TIME_WAIT 套接字，用于新的TCP连接。

> https://www.cnblogs.com/alchemystar/p/13175276.html

### 其他设置

- `net.ipv4.ip_local_reserved_ports = 30000-32767` 设置 NodePort 的保留端口范围。防止发生节点上的应用占用端口造成端口冲突。

除此以外 kubelet 和 kubeproxy 还会修改,如 `net.ipv4.ip_forward=1`等网络配置，必要设置会在 kubeadm 安装过程中进行检查，在此就不再继续展开。

## 结语

在 Kubernetes 集群中往往会运行不同种类的工作负载，这给服务器的优化带来了极大的挑战。因此在选择优化的目标时，通常是平衡而不是极致。尽可能利用 POD 与 kubelet 的设置进行POD to POD, POD to NODE 的资源限制与隔离。通用设置可参考上文中给出的推荐设置，并结合本身的硬件资源进行优化。

代码参考：

https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/component-helpers/node/util/sysctl/sysctl.go
https://github.com/kubernetes/kops/blob/master/nodeup/pkg/model/sysctls.go
https://github.com/kubesphere/kubekey/blob/master/pkg/bootstrap/os/templates/init_script.go