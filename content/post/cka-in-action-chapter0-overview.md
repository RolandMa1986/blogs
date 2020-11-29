---
title: "CKA 实战  考试概览"
date: 2020-11-18T13:22:56Z
draft: true
tag: 
- CKA
- Kubernetes
---

CKA 全称 Certified Kubernetes Administrator，是由 CNCF 联合 Linux 基金会推出官方认证项目。Kubernetes 作为当前全球最火热的云容器管理平台，拥有 CKA 证书无疑会提高自身价值。准备报考 CKA 的朋友可以跟随本系列文章一起复习考试内容。本系列文章会结合 CKA 考点，实验手册以及部分考题，带您全面学习 Kubernetes 各个知识点。

> 本系列文章参考 2020年10月的考试大纲。

## 考试范围

CKA 考试主要考察 Kubernetes 管理员最常用的工作技能，考试被划分为 5 个部分，每个部分权重如下：


领域| 权重 
---|---
集群架构，安装和配置| 25%
工作负载与调度| 15%
服务与网络|20%
存储|10%
故障排除|30%

下面我们对以上五个部分的考点逐一解析


## 集群架构，安装和配置

- 配置 Kubernetes 集群基础环境
- 使用 Kubeadm 安装集群
- 使用  Kubeadm 升级 Kubernetes 版本
- etcd 备份与恢复
- 集群高可用性管理
- 基于角色的访问控制 (RBAC)


## 工作负载与调度

- 理解 Deployments 并可以执行滚动升级和回退
- 应用缩放
- 理解 Pod 调度及资源限制对调度影响
- 使用 ConfigMaps 和 Secrets 配置应用
- Understand the primitives used to create robust, self-healing, application deployments
- Awareness of manifest management and common templating tools

## 服务与网络

- 理解集群 Node 网络配置
- 理解 Pods 网络连接
- 理解 ClusterIP, NodePort, LoadBalancer 服务类型和 endpoints
- Know how to use Ingress controllers and Ingress resources
- Know how to configure and use CoreDNS
- Choose an appropriate container network interface plugin

## 10% - Storage

- Understand storage classes, persistent volumes
- Understand volume mode, access modes and reclaim policies for volumes
- Understand persistent volume claims primitive
- Know how to configure applications with persistent storage

##  30% - Troubleshooting

- Evaluate cluster and node logging
- Understand how to monitor applications
- 管理容器 container stdout & stderr logs
- Troubleshoot application failure
- Troubleshoot cluster component failure
- Troubleshoot networking

 
## 参考文档

- 考试大纲 https://github.com/cncf/curriculum