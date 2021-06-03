---
title: "Helm Tips"
date: 2021-05-24T06:52:16Z
---

Helm 是最常用的 Kubernetes 包管理器之一。但实际应使用时却会面对各种复杂的问题，以下内容是对常见应用场景的简要总结：

### 将 kubectl 部署的资源转为 Helm 资源

如果您的项目最初不是使用 Helm 进行资源管理且已经部署在线上运行，那么你可能需要将已有资源转换 Helm 资源。 否则直接运行 helm install 这时你将看见如下错误信息：

`Error: rendered manifests contain a resource that already exists. Unable to continue with install: `

解决这个问题的方法很简单，只需要将对应的资源打上 Helm 的 Label 和 Annotation 即可:
```bash
KIND=deployment
NAME=my-app-staging
RELEASE=staging
NAMESPACE=default
kubectl annotate $KIND $NAME meta.helm.sh/release-name=$RELEASE
kubectl annotate $KIND $NAME meta.helm.sh/release-namespace=$NAMESPACE
kubectl label $KIND $NAME app.kubernetes.io/managed-by=Helm
```

注意： 第一个版本 Helm chart 中最好与原有资源一致，请勿更改 immutible field 或增加删减其他属性。

### 不要使用 Helm 以外的工具修改资源

Helm 与 kubectl 一样会使用 3-way merge 策略。使用 3-way merge 策略时，需要使用最后一次部署的资源清单和当前即将部署的资源清单进行比较，生成增删改的 Patch 操作。kubectl 会将最后一次 apply 的资源保存在 `kubectl.kubernetes.io/last-applied-configuration` 注解中。而 Helm 默认使用 Secret 保存。当使用 kubectl 修改了 Helm 所部署的资源后，可能因为无法正确的生成Patch 操作而导致升级操作失败。

### 避免 --force 强制更新

在使用 Helm 更新 资源时，应避免使用 --force 选项。 --force 的行为根据 Helm 版本的不同，行为也发生了变化。我们可以创建一个示例 Chart `your-chart`。这个chart 中包含有一个 Service 资源如下：

```yaml
apiVersion: v1
kind: Service
metadata:
  name: force-update
spec:
  ports:
  - name: https
    port: 8443
    protocol: TCP
    targetPort: https
  selector:
    app: demo
  sessionAffinity: None
  type: ClusterIP
```

接下来，如果我们使用 Helm v3.5.x 进行安装并升级此 Chart 时， 我们会遇见如下的错误信息：

```bash
helm install test ./your-chart/ --force
helm upgrade test ./your-chart/ --force
Error: UPGRADE FAILED: failed to replace object: Service "force-update" is invalid: spec.clusterIPs[0]: Invalid value: []string(nil): primary clusterIP can not be unset
```
> tip: 上述使用 Helm v3.5.2 进行测试。如果使用 Helm v3.2.1 上面更新可以成功。

对比以上行为后，我们找到了 [helm/helm#8000](https://github.com/helm/helm/pull/8000) 的改动。可以看出，在这个PR之前，如果当前资源与上一次版本一致的情况下，即使使用了 --force 选项，更新资源操作也会被跳过。而在这个 PR 之后，虽然执行了强制更新操作，但由于使用了 API 调用使用了 Replace 行为，而这个Service 却由于缺少必要的字段而更新失败。

与之对比的是 kubectl apply --force 的行为，它使用了 delete/create 的策略，保证了资源的可以被强制更新。

### 初始化密钥资源

第一次安装 helm 时进行初始化密钥资源是比较常见的一个操作，它要求在更新资源时可以保持密钥不变或者保留用户的设置。这时 lookup 函数就可以派上用场了。 lookup 函数可以在当前集群中查找资源： 一个典型的生成/使用已有密钥的操作如下：

```yaml
{{- $secret := (lookup "v1" "Secret" .Release.Namespace "some-secret" -}}
apiVersion: v1
kind: Secret
metadata:
  name: some-secret
type: Opaque

{{ if $secret -}}
data:
  apiKey: {{ $secret.data.apiKey }}

{{ else -}}
stringData:
  apiKey: {{ randAlphaNum 8 | quote }}
{{ end }}
```