---
title: "CKA 实战 第二章 安装与升级"
date: 2020-11-19T09:30:43Z
tag: 
- CKA
- Kubernetes
- kubeadm
---

Kubernetes 有多种部署方式，学习环境可以选用 Minikube 或 Kind 等。但在 CKA 认证考试中会考察基于 Kubeadm 的部署方式， 操作系统为 Ubuntu。你可以使用云主机或者VirlualBox虚拟机环境安装 Ubuntu。这里我们推荐在线教学环境 [KataKoda](https://www.katacoda.com/courses/kubernetes/getting-started-with-kubeadm#)。 Kadakda 的课程中提供了交互式教学环境，我们既可以参考其实验步骤，也可独立使用其在线环境。

## 什么是 Kubeadm?

Kubeadm 是 Kubernetes 集群快速安装部署工具，它解决了 Kubernetes 核心组件安装，TLS 加密配置，证书管理等问题，提供了开箱即用的安全特性。

### Kubeadm 常用命令

1. **kubeadm init** 用于初始化 Kubernetes 控制面
2. **kubeadm join** 将工作节点加入集群或者添加其他控制面节点
3. **kubeadm upgrade** 用户升级 Kubernetes 集群
4. **kubeadm reset** 重置当前主机，恢复 kubeadm init 或 kubeadm join 的配置。

## Kubernetes 组件

Kubernetes 集群由**控制面**与**集群节点**两部分组成， **节点**上运行应用的工作负载 Pod，**控制面**负责集群中的工作节点与 Pod 的管理。

![Kubernetes](https://d33wubrfki0l68.cloudfront.net/2475489eaf20163ec0f54ddc1d92aa8d4c87c96b/e7c81/images/docs/components-of-kubernetes.svg)

如上图所示，控制面运行了 kube-apiserver, etcd, kube-scheduler, kube-controller-manager 等组件组成。 节点组件包括 kubelet, kube-proxy, Container runtime(默认 Docker)等，他们运行在所有的节点上，包括控制面节点。

Kubernetes 集群安装过程中 Kubeadm 会负责除 kubelet, Container runtime 外的其他组件。kubelet, Container runtime 以及 kubectl 客户端需要使用 APT 安装。

除以上组件外，还需要手动安装网络插件，使集群中的 Pod 可以跨主机通讯。在集群安装前需要规划集群网络，如本实验中使用的 Calico， 其默认 Pod 网段为 `192.168.0.0/16`, 不能与主机网段冲突。使用 `kubeadm init` 命令时，需指定 `--pod-network-cidr` 参数，使其与网络插件网段匹配。

## Kubernetes 集群搭建

以下过程中我们会安装一个 Kubernetes 1.18.1 的集群，在稍后的升级实验中会将它升级为 1.19.0

### 实验环境

点击 [KataKoda](https://www.katacoda.com/courses/kubernetes/getting-started-with-kubeadm#)，即可进入实验课程。

主机名|IP|角色
---|---|---
controlplane| *172.17.0.57/16 | 控制面 
node01| *172.17.0.73/16 | 工作节点

> 注意，IP 有可能会不同。

### 安装过程

1. 添加kubernetes源：(Katakoda中可以忽略) 

```bash
sudo vim /etc/apt/sources.list.d/kubernetes.list
```

在文件中添加以下行：
```bash
deb http://apt.kubernetes.io/ kubernetes-xenial main
```
添加安装包的 GPG 密钥
```bash
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
```

2. 检查是否添加成功 (controlplane+node01均执行该步骤)

```bash 
sudo cat /etc/apt/sources.list.d/kubernetes.list
```

3. 安装kubelet, kubeadm, kubectl, (controlplane+node01均执行该步骤)

```bash 
sudo apt update
sudo apt install -y kubelet=1.18.1-00 kubeadm=1.18.1-00 kubectl=1.18.1-00
sudo apt-mark hold kubelet kubeadm kubectl
```
注意安装与升级过程中会多次使用 `apt-mark` 命令用于标记该软件包不被自动更新。

4. 在controlplane上执行集群初始化命令 kubeadm init, 初始化支持命令行参数模式和 yaml 配置格式等两种模式

4.1 使用命令直接初始化:

```bash 
sudo kubeadm init  --kubernetes-version v1.18.1 --pod-network-cidr=192.168.0.0/16
```

安装过程中，请注意控制台的输出，在输出结果中有我们下一步中需要使用到的命令：

```
[init] Using Kubernetes version: v1.18.1
[preflight] Running pre-flight checks
[WARNING IsDockerSystemdCheck]: detected "cgroupfs" as the
Docker cgroup driver. The recommended driver is "systemd".
....
you can now join any number of the control-plane node
running the following command on each as root:

kubeadm join k8smaster:6443 --token b0x4dv.nbut63ktiaikcc24 \
 --discovery-token-ca-cert-hash sha256:551fe78b50dfe52410869685b7dc70b9a27e550241a6112d8d1fef2073759bb4 \
 --control-plane --certificate-key 911d41fcada89a18210489afaa036cd8e192b1f122ebb1b79cce1818f642fab8

Please note that the certificate-key gives access to cluster sensitive
data, keep it secret!
As a safeguard, uploaded-certs will be deleted in two hours; If
necessary, you can use
"kubeadm init phase upload-certs --upload-certs" to reload certs afterward.

Then you can join any number of worker nodes by running the following
on each as root:

kubeadm join k8smaster:6443 --token b0x4dv.nbut63ktiaikcc24 \
 --discovery-token-ca-cert-hash sha256:551fe78b50dfe52410869685b7dc70b9a27e550241a6112d8d1fef2073759bb4
```

4.2 使用 yaml 配置先保存以下配置到 kubeadm-config.yaml

```bash
vim kubeadm-config.yaml
```
复制并粘贴以下 yaml 文件到编辑器中
```yaml
apiVersion: kubeadm.k8s.io/v1beta2
kind: ClusterConfiguration
kubernetesVersion: 1.18.1 #<-- Use the word stable for newest version
controlPlaneEndpoint: "controlplane:6443" #<-- Use the node alias not the IP
imageRepository: registry.aliyuncs.com/google_containers # For China user
networking: 
  podSubnet: 192.168.0.0/16 #<-- Match the IP range from the Calico config file
```

保存 yaml 文件后执行:

```bash
kubeadm init --config=kubeadm-config.yaml --upload-certs | tee kubeadm-init.out
```

5. 复制 kubeconfig 到当前用户的 .kube 目录, 用于配置 kubectl

```bash
mkdir -p $HOME/.kube
cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
chown $(id -u):$(id -g) $HOME/.kube/config
```

复制成功后既可以使用 kubectl 命令查看 Pod 状态
```bash
kubectl get pods --all-namespaces 
```

这里推荐安装 bash auto-completion 插件

```bash
# 如果尚未安装 bash-completion 执行
sudo apt-get install bash-completion -y

source <(kubectl completion bash)
echo "source <(kubectl completion bash)" >> $HOME/.bashrc
```

6. 安装配置 Calico 网络组件

```bash
kubectl apply -f  https://docs.projectcalico.org/manifests/calico.yaml
```

> 更多网络组件可以在官网中查找：
> https://kubernetes.io/docs/concepts/cluster-administration/addons/

7. 默认情况下，由于安全原因，集群并不会将pods部署在Master节点上。但是在开发环境下，我们可能就只有一个Master节点，这时可以使用下面的命令来解除这个限制：
```bash
kubectl taint nodes --all node-role.kubernetes.io/master-
```
注意命令中的 "-" 减号，它代表删除

8. 复制步骤 4 中输出的 Join 命令，并在 node01 上执行将其加入集群
```bash
kubeadm join [controlplane]:6443 --token b0x4dv.nbut63ktiaikcc24 --discovery-token-ca-cert-hash sha256:551fe78b50dfe52410869685b7dc70b9a27e550241a6112d8d1fef2073759bb4
```

注意，token 的有效期为 2 小时，如果超过两个小时，需要创建一个新的 token。

```bash
sudo kubeadm token list
sudo kubeadm token create
```

9. 在 controlplane 节点查看 node 状态

```bash
kubectl get node -o wide
```

## Kubernetes 升级

> 在生产环境升级集群前，需要对 ETCD 执行备份操作，ETCD 备份在后续章节中讲解。

### 实验环境

继续使用以上的 kubernetes 环境

### 控制面升级过程

1. 首先更新 APT 源数据，并查找当前可用版本， 本次实验中使用 1.19.0 版本
```bash
sudo apt update
sudo apt-cache madison kubeadm |grep 1.19
```
2. 使用APT 安装指定版本的 kubeadm，注意升级前需要解锁操作，升级后加锁，防止其被意外升级
```bash
sudo apt-mark unhold kubeadm
sudo apt-get install -y kubeadm=1.19.0-00
sudo apt-mark hold kubeadm
```
3. 在升级控制面节点前，需要先驱逐 Pod。 DaemonSets 需要被调度到每一个 Node 上运行，如 Calico 等无法被驱逐。因此需要忽略 DaemonSets。
```bash
kubectl drain controlplane --ignore-daemonsets
```
4.  使用 kubeadm upgrade plan 命令检查当前集群状态，然后进行升级。命令执行后有可能看到更新版本的 Kubernetes。此处我们仅使用 v1.19.0即可。
```bash
sudo kubeadm upgrade plan
```

5. 执行升级命令,升级过程会对 kube-apiserver, kube-controller-manager, kube-scheduler, kube-proxy, CoreDNS, etcd等组件进行升级。Kubelet 需要在控制面升级结束后手动升级。
```bash
sudo kubeadm upgrade apply v1.19.0
```
6. 检查节点状态可以看到此时 controlplane 仍然显示为 v1.18.1 版本。

```bash
kubectl get node

NAME STATUS ROLES AGE VERSION
controlplane Ready,SchedulingDisabled master 7h48m v1.18.1
node01       Ready                    <none> 7h46m v1.18.1
```

7. 继续升级 kubelet, kubectl 
```bash
sudo apt-mark unhold kubelet kubectl
sudo apt-get install -y kubelet=1.19.0-00 kubectl=1.19.0-00
sudo apt-mark hold kubelet kubectl
```

8. 升级结束后需要重启 kubelet 服务
```bash
 sudo systemctl daemon-reload
 sudo systemctl restart kubelet
```

10. 如果再次检查 controlplane 节点状态，我们会发现当前版本已经升级为 v1.19.0, 但是节点状态依然为 SchedulingDisabled。执行 kubectl uncordon 命令，是其可以被调度：
```bash
kubectl uncordon controlplane
```

至此控制面节点已经升级完毕

### 工作节点升级过程

工作节点升级与控制面节点升级过程基本一致， 我们继续执行工作节点的升级。

1. 在 node01 上使用 APT 安装 v1.19.0 版本的 kubeadm
```bash
sudo apt-mark unhold kubeadm
sudo apt-get update && sudo apt-get install -y kubeadm=1.19.0-00
sudo apt-mark hold kubeadm
```
2. 驱逐Pod
注意，当前 node01 上没有 kubeconfig 需要再 contolpanle 上执行， 当然也可以将 contolpanle 的 kubeconfig 复制到 node01 上
```bash
kubectl drain node01 --ignore-daemonsets
```
3. 执行节点升级命令

```bash
sudo kubeadm upgrade node
```
4. 节点升级成功后，继续升级 kubelet
```bash
sudo apt-mark unhold kubelet kubectl
sudo apt-get install -y kubelet=1.19.0-00 kubectl=1.19.0-00
sudo apt-mark hold kubelet kubectl
```
5. 重启 kubelet 服务
```bash
sudo systemctl daemon-reload
sudo systemctl restart kubelet
```
6. 检查 node 状态，此时 node01 已经升级为 v1.19.0
```bash
kubectl get node
```
7. 执行 kubectl uncordon 命令，是其可以被调度
```bash
kubectl uncordon node01
```

至此，整个集群安装与升级完成。

## 真题解析

Task

Given an existing Kubernetes cluster running version 1.18.8, upgrade all fo the Kubernetes control plane and node components on the master node only to version 1.19.0

You are also expected to upgrade kubelet and kubectl on the master node.

> Be sure to drain the master node before upgrading it and uncordon it after the upgrade. Do not upgrade the work nodes, etcd, the container manager, the CNI plugin, the DNS service or any other addons.

注意以上题目，
 - 首先，不需要升级工作节点， 因此只需要在控制面节点上执行升级操作。
 - 其次，不需要升级 etcd, DNS service, container manager, CNI plugin 等组件

> 此处有疑问，kubeadm 中只提供了 --etcd-upgrade=false 参数用于忽略 etcd 升级。未找到如何忽略其它 addon 的命令参数。
> 有考过 CKA 的朋友说只要不升级 工作节点即可。此处有疑问，仍需确认。

执行过查看升级计划命令后，我们可以看到 CoreDNS 与 etcd 在升级范围内。
```bash
sudo kubeadm upgrade plan
```

```
COMPONENT CURRENT AVAILABLE
kube-apiserver v1.18.1 v1.19.0
kube-controller-manager v1.18.1 v1.19.0
kube-scheduler v1.18.1 v1.19.0
kube-proxy v1.18.1 v1.19.0
CoreDNS 1.6.7 1.7.0
etcd 3.4.3-0 3.4.9-1
```

```bash
sudo kubeadm upgrade apply v1.19.0 --etcd-upgrade=false
```
注意

## 参考文档

- [安装 Kubeadm](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/)
- [创建 Kubernetes 集群](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/)
- [升级 Kubernetes](https://kubernetes.io/docs/tasks/administer-cluster/kubeadm/kubeadm-upgrade/)
- [安装 Calico 网络插件](https://docs.projectcalico.org/getting-started/kubernetes/self-managed-onprem/onpremises)
