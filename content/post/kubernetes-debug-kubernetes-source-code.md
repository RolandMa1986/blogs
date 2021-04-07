---
title: "Kubernetes 源码调试"
date: 2021-01-17T12:41:52Z
tag: 
- golang
- vscode
- minikube
---


## 前言

本文提供了一种基于 vscode 和 minikube 的快速搭建 Kubernetes 开发调试环境的方法。通过 devle 对 Kubernetes 各个组件进行远程调试。调试代码可以帮助你更深入地理解 Kubernetes 逻辑，更清晰的分析代码执行流程。

## 准备阶段

1. Clone kubernetes 源代码
参考官方文档，将 Kubernetes clone 到默认 $GOPATH (Ubuntu 默认路径: `~/go`)下

```bash
mkdir -p $GOPATH/src/k8s.io
cd $GOPATH/src/k8s.io
git clone https://github.com/kubernetes/kubernetes
cd kubernetes

# 以 1.19.x 分支为例我们签出源代码
git checkout release-1.19
# 或签出指定 tag 版本
git checkout v1.19.3
```
Kubernetes 各个版本的编译对 golang 版本有依赖，例如编译 Kubernetes 1.19.0 及以上版本需要 golang 1.15.0 以上版本。

2. Vscode
Vscode 安装方法请参考官网：
https://code.visualstudio.com/

Vscode 安装成功后，还需要安装 GO 扩展。成功安装 go 扩展后，使用快捷键 `Ctrl + Shift + P`， 打开 *Command Palette ...*
执行 `Go: Install/Update Tools` 命令，并选择 dlv 等插件。插件会被安装到默认的 `$GOPATH/bin` 目录中。

3. Minikube
Minikube 是最常用的搭建本地kubernetes集群工具之一。可以用于学习以及本地开发环境。 我们使用以下命令启动一个 kubernets v1.19.3 的集群实例：

```bash
minikube start --cpus=4 --memory=4096 \
--image-mirror-country cn \
--registry-mirror=https://a3dw7d3s.mirror.aliyuncs.com  \
--kubernetes-version=v1.19.3
```

下载，搭建 minikube 方法可参考：
https://developer.aliyun.com/article/221687


## 手动启动调试

动手调试前，我们需要了解一些 Kubernetes 组件启动的相关知识。Kubernetes 控制面组件中主要有三种启动方式：

- Systemd: 以 kubelet 为代表，使用 systemed 控制其生命周期，跟随系统一起启动
- Static POD: kube-apiserver, kube-controller-manager 为代表，使用静态 POD 方式，由 kubelet 控制其启动
- DaemonSet、 Deployment: CoreDNS,Storage provisioner 等插件一般以此方式启动。

### Kube-apiserver 调试

下面我们以 Kube-apiserver 为例，演示静态 POD 组件的调试的整个过程。

1. 首先我们需要编译带有调试信息的 kube-apiserver 二进制文件：

```bash
make all WHAT=cmd/kube-apiserver GOLDFLAGS=''"
```
由于 Kubernets  完全的编译所需时间很长，因此可以使用 `WHAT` 参数选择需要编译的二进制。并且使用 `GOLDFLAGS=''`
覆盖其默认的参数： -s disable symbol table, -w disable DWARF generation
执行成功后，我们即可在 _output/bin/ 目录下找到所需的二进制文件。

2. 复制 kube-apiserver 与 dlv 到 Minikube 中：

以下命令可能跟 minikube 的启动方式有关，我们以 docker 方式为例，首先获取 minikube 的 ip 地址，并将所需的文件复制到容器中

```bash
export MINIKUBE_IP=$(minikube ip)

scp -i ~/.minikube/machines/minikube/id_rsa _output/bin/kube-apiserver docker@$MINIKUBE_IP:

scp -i ~/.minikube/machines/minikube/id_rsa ~/go/bin/dlv docker@$MINIKUBE_IP:
```

3. 查看 kube-apiserver 启动参数, 并停止

以下命令需要在 minikube 容器中执行。通过查看 kube-apiserver 的 manifest 文件可以获取 kube-apiserver 启动的参数

```bash
minikube ssh sudo cat /etc/kubernetes/manifests/kube-apiserver.yaml
```

```yaml
...
spec:
  containers:
  - command:
    - kube-apiserver
    - --advertise-address=192.168.49.2
    - --allow-privileged=true
    - --authorization-mode=Node,RBAC
    - --client-ca-file=/var/lib/minikube/certs/ca.crt
... 省略
```

在启动我们自己的 kube-apiserver 前，首先需要将系统中的 kube-apiserver停止。将 kube-apiserver.yaml 从 manifest 目录移除，静态 POD 即可被 kubelet 停止。

```bash
minikube ssh 'sudo mv /etc/kubernetes/manifests/kube-apiserver.yaml /etc/kubernetes/kube-apiserver.yaml'
```

4. dlv 方式启动 kube-apiserver

启动 dlv 远程模式后，dlv 会等待 IDE 连接。注意,此处必须指定 API 版本 `--api-version=2`， 在 "--" 之后可以输入您自己所需的kube-apiserver 参数：

```bash 
minikube ssh 'sudo /home/docker/dlv exec /home/docker/kube-apiserver --headless --listen=:2345 --log --api-version=2 --  --advertise-address=1.2.3.4 ...' 
```

5. Vscode 连接远程调试连接 dlv

启动 Vscode 后打开 kubernetes 目录。选择 "run"-> "Add Configuration" 菜单后在 `launch.json` 的 configurations 节点下添加以下配置：

```json
{
    "name": "Connect to Kube-apiserver",
    "type": "go",
    "request": "attach",
    "mode": "remote",
    "remotePath": "${workspaceFolder}",
    "port": 2345,
    "host": "192.168.49.2"
}
```

接下来在 main 函数中增加一个断点，启动调试即可成功捕获到断点了。

### kubelet 调试

kubelet 调试过程与上面过程相似，我们只需要先得到 kubelet 的启动参数，然后停止 kubelet 服务即可。 我们先略去 kubelet 构建，复制等过程。直接获取其启动参数：

1. 使用 systemctl 获取 kubelet 的启动参数

```bash
$ minikube ssh sudo systemctl status kubelet
...
kubelet --bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --config=/var/lib/kubelet/config.yaml --container-runtime=docker --hostname-override=minikube --kubeconfig=/etc/kubernetes/kubelet.conf --node-ip=192.168.49.2 --pod-infra-container-image=registry.cn-hangzhou.aliyuncs.com/google_containers/pause:3.2
...
```
2. 停止 kubelet 服务
```bash
minikube ssh sudo systemctl stop kubelet
```

## 自动化过程

以上过程并不复杂，但是我们没有必要重复这些手动过程。通过 Vscode 的自定义task 我们可以自动化整个过程。

1. 首先，我们将上面的每个独立的步骤定义为一个 `shell` 命令。这个样我们就可以通过 Terminal -> Run Task 执行这些步骤了。
```json
{
    "label": "Copy Kube-apiserver",
    "type": "shell",
    "command": "scp -i ~/.minikube/machines/minikube/id_rsa _output/bin/kube-apiserver docker@192.168.49.2:"
}
```

2. 然后，我们使用一个组合命令，将上面定义的步骤按照顺序执行。
```json
        {
            "label": "Launch Kube-apiserver Debuger",
            "dependsOrder": "sequence",
            "dependsOn": ["build Kube-apiserver", "Copy Kube-apiserver", "Copy Kube-apiserver","Stop kube-apiserver","Start Kube-apiserver Debuger"]
        }
```

3. 接下来，在执行调试任务前，我们只需要执行 `Launch Kube-apiserver Debuger` 任务即可在 minikube 中启动调试环境。


完整的task.json定义参考如下：
```json
{
    "version": "2.0.0",
    "tasks": [
        {
            "label": "Init dlv",
            "type": "shell",
            "command": "scp -i ~/.minikube/machines/minikube/id_rsa ~/go/bin/dlv docker@192.168.49.2:"
        },
        {
            "label": "Copy Kube-apiserver",
            "type": "shell",
            "command": "scp -i ~/.minikube/machines/minikube/id_rsa _output/bin/kube-apiserver docker@192.168.49.2:"
        },
        {
            "label": "build Kube-apiserver",
            "type": "shell",
            "command": "make all WHAT=cmd/kube-apiserver GOLDFLAGS=''",
            "group": "build",
            "problemMatcher": [
                "$go"
            ]
        },
        {
            "label": "Stop kube-apiserver",
            "type": "shell",
            "command": "minikube ssh 'if [ -f /etc/kubernetes/manifests/kube-apiserver.yaml ];then sudo mv /etc/kubernetes/manifests/kube-apiserver.yaml /etc/kubernetes/kube-apiserver.yaml;fi'"
        },
        {
            "label": "Start Kube-apiserver Debuger",
            "type": "shell",
            "command": "minikube ssh 'sudo /home/docker/dlv exec /home/docker/kube-apiserver --headless --listen=:2345 --log --api-version=2 -- -v=4 --advertise-address=192.168.49.2 --allow-privileged=true --authorization-mode=Node,RBAC --client-ca-file=/var/lib/minikube/certs/ca.crt --enable-admission-plugins=NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota --enable-bootstrap-token-auth=true --etcd-cafile=/var/lib/minikube/certs/etcd/ca.crt --etcd-certfile=/var/lib/minikube/certs/apiserver-etcd-client.crt --etcd-keyfile=/var/lib/minikube/certs/apiserver-etcd-client.key --etcd-servers=https://192.168.49.2:2379 --insecure-port=0 --kubelet-client-certificate=/var/lib/minikube/certs/apiserver-kubelet-client.crt --kubelet-client-key=/var/lib/minikube/certs/apiserver-kubelet-client.key --kubelet-preferred-address-types=InternalIP,ExternalIP,Hostname --proxy-client-cert-file=/var/lib/minikube/certs/front-proxy-client.crt --proxy-client-key-file=/var/lib/minikube/certs/front-proxy-client.key --requestheader-allowed-names=front-proxy-client --requestheader-client-ca-file=/var/lib/minikube/certs/front-proxy-ca.crt --requestheader-extra-headers-prefix=X-Remote-Extra- --requestheader-group-headers=X-Remote-Group --requestheader-username-headers=X-Remote-User --secure-port=8443 --service-account-key-file=/var/lib/minikube/certs/sa.pub --service-cluster-ip-range=10.96.0.0/12 --tls-cert-file=/var/lib/minikube/certs/apiserver.crt --tls-private-key-file=/var/lib/minikube/certs/apiserver.key'"
        },
        {
            "label": "Launch Kube-apiserver Debuger",
            "dependsOrder": "sequence",
            "dependsOn": ["build Kube-apiserver", "Copy Kube-apiserver", "Copy Kube-apiserver","Stop kube-apiserver","Start Kube-apiserver Debuger"]
        }
    ]
}
```

## 最后

以上过程使用了 Vscode, Minikube 等工具快速搭建了一个 kubernetes 源码开发调试环境。除以上基本方法外，我们可以进一步自动化整个过程。例如，利用 preLaunchTask 属性和 problemMatcher 属性配置一键 debug。也可以使用变量替换 minikube 的 ip地址。基于以上方法，我们也可以快速切换 kubernetes 版本而无需再本地或远程执行整个 kubernetes 搭建过程。工欲善其事必先利其器，希望以上方法对您的 Kubernetes 源码学习过程中有所帮助。