---
title: "基于Ubuntu18.04的云原生应用开发环境搭建 – KinD 搭建Kubernetes多节点集群"
date: 2020-10-19T01:13:57Z
---

# 基于Ubuntu18.04的云原生应用开发环境搭建 -- KinD 搭建Kubernetes多节点集群

KinD是Kubernetes in Docker的简称，它的主要目标是用于测试Kubernetes本身。也可以用于搭建本地Kubernetes开发环境，开发云原生应用。KinD可以运行在Windows，Macos和linux中。本文以Ubuntu为例，介绍kind的常用命令以及配置。

## 前提条件

- 已安装Ubuntu18.04桌面版
- 已安装Docker

## 下载最新版KinD

当前kind版本为v0.9.0,可以使用以下命令下载并安装

```bash
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.9.0/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/
```

## Kind使用

### 创建集群

创建Kubernetes集群最简单的方式就是使用`kind create cluster`命令。执行后3-5分钟即可启动一个K8S集群。同时kind会更新`${HOME}/.kube/config`下的kubeconfig。默认集群名字为kind，也可通过--name 指定名称。

### 常用命令

- 查看已创建的集群

```bash
kind get clusters
```

- 删除集群

```bash
kind delete cluster --name your-cluster
```

## 高级选项

### 多节点集群

kind最重要的功能之一是创建多节点集群，为了创建多个node，我们需要先创建一个yaml配置文件。如：

```yaml
# three node (two workers) cluster config
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
- role: worker
- role: worker
```
然后通过--config 选项创建集群。

### 高级配置

日常工作中， 有以下三个常用配置供大家参考:
1. 通过image配置Kubernetes版本为1.17.5
2. 将/etc/localtime映射到node的容器。注意:很多helm chart中会使用到localtime。但是node的容器中并不包含localtime，因此挂载。
3. 使用镜像加速docker.io访问。注意，kind node中默认使用cri-o容器引擎。

```
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  image: kindest/node:v1.17.5
  extraMounts:
  - hostPath: /etc/localtime
    containerPath: /etc/localtime
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
    endpoint = ["https://docker.mirrors.ustc.edu.cn"]
```
更多选项可以在官网查找:
https://kind.sigs.k8s.io/docs/user/configuration/
