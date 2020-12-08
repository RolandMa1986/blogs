---
title: "Kubernetes - 'No Podsandbox Found with Id' 导致 Kubelet 启动失败分析过程总结"
date: 2020-12-08T03:38:37Z
tag: 
- Kubernetes
---

某线上客户反应 Kubernetes 集群中一节点显示为 "NotReady" 状态，重启 Kubelet 后任然无法恢复。由于无法连接环境，只能拿到少量相关日志信息，因此尝试本地重现。

<!--more-->
## 收集信息

### 查看 Node 状态

使用 describe 命令后可以看到，该机器当前 CPU/Memery/Disk 空间充足，但是 container runtime 状态未返回。

```log
$ kubectl describe node node3
Name:               node3
Roles:              worker
...
Conditions:
  Type                 Status  LastHeartbeatTime                 LastTransitionTime                Reason            Message
  ----                 ------  -----------------                 ------------------                ------            -------
Ready                False   Mon, 07 Dec 2020 11:16:28 +0800   Mon, 07 Dec 2020 10:53:24 +0800   KubeletNotReady     container runtime status check may not have completed yet
...
Capacity:
  cpu:                2
  ephemeral-storage:  41611416Ki
  hugepages-2Mi:      0
  memory:             3880504Ki
  pods:               110
Allocatable:
  cpu:                1600m
  ephemeral-storage:  41611416Ki
  hugepages-2Mi:      0
  memory:             3250666289
  pods:               110
System Info:
  Kernel Version:               3.10.0-1160.6.1.el7.x86_64
  OS Image:                     CentOS Linux 7 (Core)
  Operating System:             linux
  Architecture:                 amd64
  Container Runtime Version:    docker://19.3.13
  Kubelet Version:              v1.18.6
  Kube-Proxy Version:           v1.18.6
```

### 日志分析

在获取集群基本信息后，怀疑是由于 Docker 导致的，但是由于一些其他原因导致未能收集到 docker 日志。

```log 
$ journalctl -xef -u kubelet
# 已去除部分不相关信息
15:16:15 node3 kubelet[14756]: I1207 15:16:15.718067   14756 server.go:417] Version: v1.18.6
...
15:16:15 node3 kubelet[14756]: I1207 15:16:15.935335   14756 client.go:75] Connecting to docker on unix:///var/run/docker.sock
15:16:15 node3 kubelet[14756]: I1207 15:16:15.935369   14756 client.go:92] Start docker client with request timeout=2m0s
15:16:16 node3 kubelet[14756]: I1207 15:16:16.099278   14756 docker_service.go:258] Docker Info: &{ID:ELYA:RGDA:NT3O:IYA7:QLCN:IBXB:QEKW:C2MQ:ITDJ:7VSA:NHI3:HXEB Containers:86 ContainersRunning:0 ContainersPaused:0 ContainersStopped:86 Images:176 
15:16:16 node3 kubelet[14756]: I1207 15:16:16.099493   14756 docker_service.go:271] Setting cgroupDriver to systemd
...
15:16:22 node3 kubelet[14756]: I1207 15:16:22.529963   14756 kubelet.go:1822] Starting kubelet main sync loop.
15:16:22 node3 kubelet[14756]: E1207 15:16:22.530073   14756 kubelet.go:1846] skipping pod synchronization - [container runtime status check may not have completed yet, PLEG is not healthy: pleg has yet to be successful]
...
15:16:23 node3 kubelet[14756]: F1207 15:16:23.071262   14756 kubelet.go:1384] Failed to start ContainerManager failed to build map of initial containers from runtime: no PodsandBox found with Id 'd0831a4288e261b5e3a5643f4a36bb2162282262fba671a682159494dfcd6f56'
15:16:23 node3 systemd[1]: kubelet.service: main process exited, code=exited, status=255/n/a
15:16:23 node3 systemd[1]: Unit kubelet.service entered failed state.
15:16:23 node3 systemd[1]: kubelet.service failed.
...
```

首先，可以确定 docker 正在运行并连接成功。Docker Info 现在当前机器上没有正在运行的容器，只有86个停止的容器。但上面日志中有两条信息引起了我们的注意：
1. PLEG is not healthy: pleg has yet to be successful
2. Failed to start ContainerManager failed to build map of initial containers from runtime: no PodsandBox found with Id xxx

*PLEG is not healthy* 是最常见的导致 NodeNotReady 的原因之一， 我们稍后再详细分析。然而导致 Kubelet 启动失败并退出的直接原因是 "Failed to start ContainerManager", 即容器管理器启动失败。

## 错误信息代码定位

使用 "no PodsandBox found with Id" 关键字，直接搜索源码我们很容易定位到位于 Container Manager 中的以下方法：

```golang
// # pkg/kubelet/cm/container_manager_linux.go
func buildContainerMapFromRuntime(runtimeService internalapi.RuntimeService) (containermap.ContainerMap, error) {
	podSandboxMap := make(map[string]string)
	podSandboxList, _ := runtimeService.ListPodSandbox(nil)
	for _, p := range podSandboxList {
		podSandboxMap[p.Id] = p.Metadata.Uid
	}

	containerMap := containermap.NewContainerMap()
	containerList, _ := runtimeService.ListContainers(nil)
	for _, c := range containerList {
		if _, exists := podSandboxMap[c.PodSandboxId]; !exists {
			return nil, fmt.Errorf("no PodsandBox found with Id '%s'", c.PodSandboxId)
		}
		containerMap.Add(podSandboxMap[c.PodSandboxId], c.Metadata.Name, c.Id)
	}

	return containerMap, nil
}

func (cm *containerManagerImpl) Start(...) ) error {
...
	// Initialize CPU manager
	if utilfeature.DefaultFeatureGate.Enabled(kubefeatures.CPUManager) {
        containerMap, err := buildContainerMapFromRuntime(runtimeService)
    ...
    }
}
```

以上代码过程比较直接：

1. 首先 Kubelet 启动过程中会调用 ContainerManager.Start() 方法，用于启动容器管理器
2. 当启用 CPU Manager 时，CPU Manager 首先获取当前 Node 上已经存在的容器(v1.8 以后默认为开启状态)
3. 由于 Kubernetes 中的 POD 是由一个 pause 容器与若干工作负载容器组成的，因此buildContainerMapFromRuntime 中分别调用了 ListPodSandbox  和 ListContainers 获取他们的键值对
4. 当一个工作负载仍然存在，但没有 pause 容器信息的抛出 `no PodsandBox found with Id xxx` 错误

了解到以上信息后，我们开始尝试重现。

## 第一次尝试：

根据以上的分析，我们只需要一个没有 pause 的容器即可。因此我们先尝试删除 pause 容器。

1. 创建一个 Deployment， 并查看POD状态
```bash
$ kubectl create deployment nginxx --image=nginx:latest
$ kubectl get pod 
NAME                           READY   STATUS    RESTARTS   AGE
nginxx-7664b4dd7-db8zt         1/1     Running   0          3m16s
```
2. 使用 docker 命令查看启动的容器， 这里我们使用了 lable 过滤条件

```bash
$ docker ps -f "label=io.kubernetes.pod.name=nginxx-7664b4dd7-db8zt"
CONTAINER ID    IMAGE   COMMAND                  CREATED             STATUS              PORTS               NAMES
3109651451ed    nginx   "/docker-entrypoint.…"   5 minutes ago       Up 5 minutes k8s_nginx_nginxx-7664b4dd7-db8zt_default_9f4c1e6b-af2e-47c3-b6ee-0b5f2d54364e_0
3a26e144d020    pause:3.2   "/pause"             6 minutes ago       Up 6 minutes k8s_POD_nginxx-7664b4dd7-db8zt_default_9f4c1e6b-af2e-47c3-b6ee-0b5f2d54364e_0
```

3. 由于 Kubelet 会不断的执行 reconcil 过程，使容器状态与目标状态保持一致。因此在删除 pause 容器时，我们先停止 kubelet 再删除。
```bash
$ sudo systemctl stop kubelet

# 删除 docker 容器
$ docker rm -f 3a26e144d020
# 重启 kubelet
$ sudo systemctl start kubelet
```

然而，并没有像我们想象中的那样简单，问题不能重现。kubelet 正常启动成功。从 docker 命令中我们可以看到，docker 将原来的容器停止，并创建一组 POD 名相同的容器：

```bash
$ docker ps -f "label=io.kubernetes.pod.name=nginxx-7664b4dd7-db8zt"
CONTAINER ID        IMAGE    COMMAND                  CREATED             STATUS                      PORTS               NAMES
9ed770770949        nginx    "/docker-entrypoint.…"   14 seconds ago      Up 13 seconds     k8s_nginx_nginxx-7664b4dd7-db8zt_default_9f4c1e6b-af2e-47c3-b6ee-0b5f2d54364e_1
a4f3dede3687        pause:3.2   "/pause"                 18 seconds ago      Up 16 seconds  k8s_POD_nginxx-7664b4dd7-db8zt_default_9f4c1e6b-af2e-47c3-b6ee-0b5f2d54364e_0
3109651451ed        nginx    "/docker-entrypoint.…"   29 minutes ago      Exited (0) 17 seconds ago 8s_nginx_nginxx-7664b4dd7-db8zt_default_9f4c1e6b-af2e-47c3-b6ee-0b5f2d54364e_0
```

## POD 中的容器如何关联？

经过一次失败的尝试后，我们继续研究容器与 POD 的关系。由于经常使用 docker， 因此第一反应是容器通过 Label 进行关联的。我们首先使用 docker inspect 查看相关容器的 Label。

```bash
$ docker inspect  9ed770770949 | jq -r '.[0].Config.Labels'
{
    ...
  "io.kubernetes.container.name": "nginx",
  "io.kubernetes.docker.type": "container",
  "io.kubernetes.pod.name": "nginxx-7664b4dd7-db8zt",
  "io.kubernetes.pod.namespace": "default",
  "io.kubernetes.pod.uid": "9f4c1e6b-af2e-47c3-b6ee-0b5f2d54364e",
  "io.kubernetes.sandbox.id": "a4f3dede368715fe288e4b079d1b2a12e0fe1f5dc1f6602ab0d06bfa4851ccdf"
  ...
}

$ docker inspect  a4f3dede3687 | jq -r '.[0].Config.Labels'
{
...
  "io.kubernetes.container.name": "POD",
  "io.kubernetes.docker.type": "podsandbox",
  "io.kubernetes.pod.name": "nginxx-7664b4dd7-db8zt",
  "io.kubernetes.pod.namespace": "default",
  "io.kubernetes.pod.uid": "9f4c1e6b-af2e-47c3-b6ee-0b5f2d54364e",
...
}
```

对于 pause 容器与工作容器，他们的 "io.kubernetes.pod.*" 属性均相同，然而对于 pause 容器，其 `io.kubernetes.docker.type` 为`podsandbox` 而工作容器为 `container`。 为进步验证我们猜想。我们找到 ListContainer 的实现。

### ListContainer 的实现

```golang
//pkg/kubelet/dockershim/docker_container.go
// ListContainers lists all containers matching the filter.
func (ds *dockerService) ListContainers(_ context.Context, r *runtimeapi.ListContainersRequest) (*runtimeapi.ListContainersResponse, error) {
	filter := r.GetFilter()
	opts := dockertypes.ContainerListOptions{All: true}

	opts.Filters = dockerfilters.NewArgs()
	f := newDockerFilter(&opts.Filters)
	// Add filter to get *only* (non-sandbox) containers.
	f.AddLabel(containerTypeLabelKey, containerTypeLabelContainer)

    ...
   
	containers, err := ds.client.ListContainers(opts)
	if err != nil {
		return nil, err
	}
	// Convert docker to runtime api containers.
	result := []*runtimeapi.Container{}
    
    ...

	return &runtimeapi.ListContainersResponse{Containers: result}, nil
}
```
ListContainers 实现比较直接，方法中开始位置设置了一个默认过滤器，用于过滤工作容器。然后调用 docker client 调用 docker api 获取容器。 最后将 docker 对象转换为 k8s 的 runtimeapi.Container 对象。

```golang
	containerTypeLabelKey       = "io.kubernetes.docker.type"
	containerTypeLabelContainer = "container"
    f.AddLabel(containerTypeLabelKey, containerTypeLabelContainer)
```

### ListPodSandbox 的实现

接着看 ListPodSandbox 的实现。首先，主体结构基本一致，首先添加默认过滤器，仅查询 `podsandbox` 类型的容器。然而除查询 docker api 获取 pause (PodSandbox) 容器信息外， 我们同时注意到 ListPodSandbox 会通过 CheckpointManager 获取容器信息。

```golang
// ListPodSandbox returns a list of Sandbox.
func (ds *dockerService) ListPodSandbox(_ context.Context, r *runtimeapi.ListPodSandboxRequest) (*runtimeapi.ListPodSandboxResponse, error) {
	filter := r.GetFilter()

    ...

	f.AddLabel(containerTypeLabelKey, containerTypeLabelSandbox)

	// Make sure we get the list of checkpoints first so that we don't include
	// new PodSandboxes that are being created right now.
	var err error
	checkpoints := []string{}
	if filter == nil {
		checkpoints, err = ds.checkpointManager.ListCheckpoints()
		if err != nil {
			klog.Errorf("Failed to list checkpoints: %v", err)
		}
	}

	containers, err := ds.client.ListContainers(opts)
	if err != nil {
		return nil, err
	}

	// Convert docker containers to runtime api sandboxes.
	result := []*runtimeapi.PodSandbox{}
	// using map as set
	sandboxIDs := make(map[string]bool)
	for i := range containers {...
	}

	// Include sandbox that could only be found with its checkpoint if no filter is applied
	// These PodSandbox will only include PodSandboxID, Name, Namespace.
	// These PodSandbox will be in PodSandboxState_SANDBOX_NOTREADY state.
	for _, id := range checkpoints {
		if _, ok := sandboxIDs[id]; ok {
			continue
		}
		checkpoint := NewPodSandboxCheckpoint("", "", &CheckpointData{})
		err := ds.checkpointManager.GetCheckpoint(id, checkpoint)
		if err != nil {...
		}
		result = append(result, checkpointToRuntimeAPISandbox(id, checkpoint))
	}

	return &runtimeapi.ListPodSandboxResponse{Items: result}, nil
}
```

根据设计文档 [CRI: Dockershim PodSandbox Checkpoint](https://github.com/freehan/community/blob/3c1ab686ad29da4f2fa900a2e31795eadd1987d4/contributors/design-proposals/cri-dockershim-checkpoint.md) 我们得知， Kubelet 仅在容器创建过程中将容器配置传递给容器运行时，因此创建 Sandbox 时 dockershim 中会将配置信息 checkpoint (/var/lib/dockershim/sandbox/) 文件中。即使 PodSandbox 容器被删除也可以从 checkpoints 中恢复出 PodSandbox 的信息。 因此，ListPodSandbox 在读取过 docker 中的容器信息后，又尝试会尝试从 checkpoint 中恢复。

## 第二次尝试：

经过上面的分析，我们发现不仅仅需要删除 pause 容器，还需要同时删除 checkpoint 才可能重现上面问题。 Deploy 过程同上, 我们从第二步开始：

1. 首先获取 POD 的名字，然后根据 POD 找到所需的容器, 并查询 sandbox pod的id：

```bash
$ docker ps -a -f "label=io.kubernetes.pod.name=nginx-one-77fbc7959f-xcmvw"
CONTAINER ID        IMAGE     COMMAND                  CREATED             STATUS              PORTS               NAMES
b9f3f73c3e68        nginx     "nginx -g 'daemon of…"   3 hours ago         Up 3 hours  k8s_nginx_nginx-one-77fbc7959f-xcmvw_accounting_c53bd2fd-cd78-43d3-b84a-441993614902_0
bd7408eadf1c        pause:3.2 "/pause"                 3 hours ago         Up 3 hours  k8s_POD_nginx-one-77fbc7959f-xcmvw_accounting_c53bd2fd-cd78-43d3-b84a-441993614902_0
```
PodSanbox 容器中并未保存 sandbox的ID，需要将工作容器的 id 传入 docker inspect 命令用于获取 sandbox id：
```bash
$ docker inspect  b9f3f73c3e68 | jq -r '.[0].Config.Labels["io.kubernetes.sandbox.id"]'

bd7408eadf1c2b859106d730fc29c3aa45f0721b17c77d0fa0d302ef1abac006
```
2. 得到以上信息后，我们就可以将它删除了：

```bash
docker rm -f bd7408eadf1c
sudo rm /var/lib/dockershim/sandbox/bd7408eadf1c2b859106d730fc29c3aa45f0721b17c77d0fa0d302ef1abac006
```

再次启动 Kubelet，我们得到了相同的启动错误：

```log
F1208 09:11:53.053603   19320 kubelet.go:1386] Failed to start ContainerManager failed to build map of initial containers from runtime: no PodsandBox found with Id 'bd7408eadf1c2b859106d730fc29c3aa45f0721b17c77d0fa0d302ef1abac006'
```

## 未完待续。。。

虽然我们已经通过手动制造脏数据的方式成功的重现了同样的错误，但是尚未找到重现真正问题的方法。这里有个疑问: 工作容器是先于 pause 容器删除的， 那么在什么情况下会导致工作容器未被删除，而 pause 容器和 checkpoints 均被删除的。 由于未能得到问题发生时的log， 因此未能进一步分析。
