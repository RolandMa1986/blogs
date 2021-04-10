---
title: "Kubernetes CNI 插件性能测试"
date: 2021-04-07T01:13:57Z
---

网络性能是影响 Kubernetes 集群与应用性能的重要指标之一。同时网络架构也是 Kubernetes 中较为复杂模块之一，Kubernetes 通过开放 CNI 接口，实现了网络插件即插即用功能，CNI 插件的实现方式是影响网络性能的最主要因素。目前主流的 CNI 插件包括：Flannel、Calico、Weave、Canal 和 Cilium 等。除 CNI 插件外，Kube-proxy 运行模式也是影响网络性能的重要因素之一。 Kubernetes 通过 Kube-proxy 提供了集群内负载均衡的能力，其内置方式有 Iptables 和 IPVS 等两种。在不同的场景下，CNI 和 Kube-proxy 都会对网络性能产生显著的影响。本文着重对 Kubernetes 网络性能测试的方法以及影响因素进行讲解和分析，不包含具体插件的性能比较。

## Kubernetes 网络通讯模式
在 Kubernetes 网络中，主要有以下几种通讯场景：

1. 容器与容器之间的直接通信，通过本地回环实现，不通过网络插件协议栈，因此测试可以忽略。
2. Pod与Pod之间的通信，这是测试中的重点，同等条件(物理网卡带宽、CPU等)下性能主要受 CNI 插件影响。
3. Pod到Service之间的通信，会受到 CNI 与 Kube-proxy 的共同影响，因此也是测试的重点之一。
4. 集群外部与内部组件之间的通信。集群外部访问的模式较多，链路较长，本文不做重点介绍。

![pod to pod](https://pek3b.qingstor.com/roland-blogs/image/spring-cloud-on-kubeshpere/pod-to-pod.png)

![pod to service](https://pek3b.qingstor.com/roland-blogs/image/spring-cloud-on-kubeshpere/pod-to-service.png)




除以上场景外，还需要考虑 POD 之间通讯是在同 NODE 与不同 NODE 等两种情况。如果再加上网络协议类型，那么我们需要考虑：通讯场景、协议、跨主机等三个维度的组合。以最常见的模式，我们可以得到以下8种组合：

TCP/UDP | Pod to Pod | Pod to Service
---     |    ---     | -----
local   |     y      |      y
Remote  |     y      |      y

当然以上并未覆盖所有场景，比如： 
1. Pod 通过 Service 访问自身暴露的服务时，需要使用 Hairpin 模式。可以根据具体场景选择测试与否。
2. Host Network 即 Pod 中运行的应用程序可以直接使用宿主主机的网络接口。使用 Host Network 时，可以认为其性能只受物理资源限制，不受容器网络，插件等其他因素影响。因此可以使用此模式进行基准测试。

## 网络测试性能指标

网络性能主要两个指标是带宽和延时。延迟决定最大的QPS(Query Per Second)，而带宽决定了可支撑的最大负荷。除以上两个指标外，在 Kubenetes 集群中另一个不可忽略的因素是 CPU 与内存。高速的网络传输时不可以避免消耗更多CPU资源，而通常 POD 都会受到 CPU 资源限制，因此 CPU 和 内存消耗也应列入性能考察指标中。 

除以上两个指标外，还需对网络质量进行考察，如网络延迟抖动，数据包丢包率等。因此需对以上两个指标进行统计计算，如平均值，最大值，90%值等。

## 测试工具
qperf 和 iperf3 是最常用的网络性能测试工具，对 Kubernetes 网络测试仍需依赖这两款软件。

- iperf3 是一款基于TCP/IP和UDP/IP的网络性能测试工具，可以用来测量网络带宽和网络质量，提供网络延迟抖动、数据包丢失率、最大传输单元等统计信息。iperf3 可以设置数据统计的间隔，用来统计网络波动情况。
- qperf 可以用来测试两个节点之间的带宽（bandwidth）和延迟（latency），除此以外还可以统计 CPU 使用率。但输出结果中只包含运行时间段的平均值。

对于 Kubernetes 测试，我们也可以使用以下的脚本或工具进行测试：

- [knb](https://github.com/InfraBuilder/k8s-bench-suite)，是一个 bash 脚本，用于在目标 Kubernetes 集群启动性能测试，性能测试基于 iperf3 实现，并内置了一个系统资源监控工具。 
- [Kubernetes netperf](https://github.com/kubernetes/perf-tests/tree/master/network/benchmarks/netperf) 是官办的测试工具集，提供了非常丰富的测试组合，可以覆盖大部分场景，并提供了 iperf3 与 netperf 的测试结果。

以上工具的使用可以查看官方文档， 我们以 qperf 为例演示一下手动测试过程。
## 基于 qperf 的测试

### 部署 qperf Pod

以下脚本启动了一个 qperf-server，并监听了 4000 端口。我们使用 nodeSelector 将服务运行在 node1 节点上：

```yaml
cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: qperf-server
  namespace: default
  labels:
    app: qperf
spec:
  containers:
  - image: "arjanschaaf/centos-qperf"
    args: ["-lp", "4000"]
    imagePullPolicy: "IfNotPresent"
    name: "qperf-server"
    ports:
    - containerPort: 4000
      name: "p1udp"
      protocol: UDP
    - containerPort: 4000
      name: "p1tcp"
      protocol: TCP
    - containerPort: 4001
      name: "p2tcp"
      protocol: TCP
    - containerPort: 4001
      name: "p2udp"
      protocol: UDP
  restartPolicy: Always
  nodeSelector:
    kubernetes.io/hostname: node1
EOF
```
### 部署 qperf 服务
如果需要针对服务进行测试，我们可以使用以下脚本暴露服务。
```yaml
cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  labels:
    app: perf
spec:
  ports:
  - port: 4000
    protocol: TCP
    targetPort: 4000
  - port: 4000
    protocol: UDP
    targetPort: 4000
  - port: 4001
    protocol: TCP
    targetPort: 4001
  - port: 4001
    protocol: UDP
    targetPort: 4001
  selector:
    app: perf
  type: ClusterIP
EOF
```

### 执行测试

当我们需要测试跨节点 Pod 的带宽与延迟时，我们在命令中使用了 tcp_bw tcp_lat 等参数，并指定在 Node2 上运行 qperf-client：

```bash
# 获取 Pod IP
serverip=`kubectl get pod qperf-server -o jsonpath='{ .status.podIP }'`
kubectl run qperf-client -it --rm --image="arjanschaaf/centos-qperf" --overrides='{ "spec": { "template": { "spec": { "nodeSelector": {"kubernetes.io/hostname": "node2" } } } } }' -- -v $serverip -lp 4000 -ip 4001 -t 120 tcp_bw tcp_lat
```

测试结果输出如下： 
```
tcp_bw
    bw              =   4.04 GB/sec
    msg_rate        =   61.7 K/sec
    port            =  4,001 
    time            =    120 sec
    send_cost       =    508 ms/GB
    recv_cost       =    508 ms/GB
    send_cpus_used  =    206 % cpus
    recv_cpus_used  =    206 % cpus
tcp_lat:
    latency        =   11.9 us
    msg_rate       =   84.2 K/sec
    port           =  4,001 
    time           =    120 sec
    loc_cpus_used  =    121 % cpus
    rem_cpus_used  =    121 % cpus 
```

如需测试 Pod 到服务的网路带宽，我们仅需要替换服务 IP 参数：
```
serverip=`kubectl get svc qperf-server -o jsonpath='{ .spec.clusterIP }'`
```
## 最后
上面内容简略的分析了 Kubernetes 下网络性能测试的基本场景以及常用指标。并简要的介绍了几款网络性能测试工具。最后演示了如何使用 qperf 进行网络测试。选择合适的场景及指标是性能测试工作的重点，在网络性能的优化以及插件选择过程中都需要以上数据的支撑。