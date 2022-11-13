---
title: "Kiosk 快速入门"
date: 2021-01-02T21:24:12+08:00
draft: true
---
<!--more-->

## Kiosk 快速入门

### 0. 环境准备

#### 0.1. CLI Tools

- kubectl： 参考文档 https://kubernetes.io/docs/tasks/tools/install-kubectl/
- helm version 3： 参考文档 https://helm.sh/docs/intro/install/

#### 0.2. Kubernetes 集群

Kisok 最低支持 Kubernetes v1.14 版。

#### 0.3. Admin Context

kube-context 必须具有 admin 权限。如果以下三个命令均返回 `yes`， 你可能具有需的权限：

```bash
kubectl auth can-i "*" "*" --all-namespaces
kubectl auth can-i "*" namespace
kubectl auth can-i "*" clusterrole
kubectl auth can-i "*" crd
```

### 1. Install kiosk

使用 helm v3 进行安装 kiosk
```bash
kubectl create namespace kiosk
helm install kiosk --repo https://charts.devspace.sh/ kiosk --namespace kiosk --atomic
```
安装完成后，验证以下 POD 正常运行：

```bash
$ kubectl get pod -n kiosk

NAME                     READY   STATUS    RESTARTS   AGE
kiosk-58887d6cf6-nm4qc   2/2     Running   0          1h
```

### 2. Configure Accounts

以下步骤中，我们使用 Kubernetes 用户模拟


2.1. Create Account


```yaml account.yaml
apiVersion: tenancy.kiosk.sh/v1alpha1
kind: Account
metadata:
  name: johns-account
spec:
  subjects:
  - kind: User
    name: john
    apiGroup: rbac.authorization.k8s.io
```

### 3. Working with Spaces

#### 3.1. Allow Users To Create Spaces

```yaml rbac-creator.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kiosk-creator
subjects:
- kind: Group
  name: system:authenticated
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: kiosk-edit
  apiGroup: rbac.authorization.k8s.io
```

#### 3.2. Create Spaces

```yaml space.yaml
apiVersion: tenancy.kiosk.sh/v1alpha1
kind: Space
metadata:
  name: johns-space
spec:
  # spec.account can be omitted if the current user only belongs to a single account
  account: johns-account
```

#### 3.3. View Spaces

```bash
# List all Spaces as john:
kubectl get spaces --as=john

# Get the defails of one of john's Spaces:
kubectl get space johns-space -o yaml --as=john
```

#### 3.4. Use Spaces

```
kubectl apply -n johns-space --as=john -f https://raw.githubusercontent.com/kubernetes/website/master/content/en/examples/application/deployment.yaml
```

#### 3.5. Create Deletable Spaces

```yaml account-deletable-spaces.yaml
apiVersion: tenancy.kiosk.sh/v1alpha1
kind: Account
metadata:
  name: johns-account-deletable-spaces
spec:
  space: 
    clusterRole: kiosk-space-admin
  subjects:
  - kind: User
    name: john
    apiGroup: rbac.authorization.k8s.io
```

```yaml space-deletable.yaml
apiVersion: tenancy.kiosk.sh/v1alpha1
kind: Space
metadata:
  name: johns-space-deletable
spec:
  account: johns-account-deletable-spaces
```

#### 3.6. Delete Spaces

```bash
kubectl get spaces --as=john
kubectl delete space johns-space-deletable --as=john
kubectl get spaces --as=john
```

#### 3.7. Defaults for Spaces

```yaml account-default-space-metadata.yaml
apiVersion: tenancy.kiosk.sh/v1alpha1
kind: Account
metadata:
  name: johns-account-default-space-metadata
spec:
  space: 
    clusterRole: kiosk-space-admin
    spaceTemplate:
      metadata:
        labels:
          some-label: "label-value"
          some--other-label: "other-label-value"
        annotations:
          "space-annotation-1": "annotation-value-1"
          "space-annotation-2": "annotation-value-2"
  subjects:
  - kind: User
    name: john
    apiGroup: rbac.authorization.k8s.io
```

### 4. Setting Account Limits

#### 4.1. Limit Number of Spaces

```yaml account-default-space-metadata.yaml
apiVersion: tenancy.kiosk.sh/v1alpha1
kind: Account
metadata:
  name: johns-account
spec:
  space:
    limit: 2
  subjects:
  - kind: User
    name: john
    apiGroup: rbac.authorization.k8s.io
```

```bash
# List existing spaces:
kubectl get spaces --as=john

# Create space-2 => should work if you had only one Space for this Account so far
kubectl apply -f https://raw.githubusercontent.com/kiosk-sh/kiosk/master/examples/space-2.yaml --as=john

# Create space-3 => should result in an error
kubectl apply -f https://raw.githubusercontent.com/kiosk-sh/kiosk/master/examples/space-3.yaml --as=john
```

#### 4.2. AccountQuotas

```yaml accountquota.yaml
apiVersion: config.kiosk.sh/v1alpha1
kind: AccountQuota
metadata:
  name: default-user-limits
spec:
  account: johns-account
  quota:
    hard:
      pods: "2"
      limits.cpu: "4"
```

### 5. Working with Templates

#### 5.1. Manifest Templates

```yaml template-manifests.yaml
apiVersion: config.kiosk.sh/v1alpha1
kind: Template
metadata:
  name: space-restrictions
resources:
  manifests:
  - kind: NetworkPolicy
    apiVersion: networking.k8s.io/v1
    metadata:
      name: deny-cross-ns-traffic
    spec:
      podSelector:
        matchLabels:
      ingress:
      - from:
        - podSelector: {}
  - apiVersion: v1
    kind: LimitRange
    metadata:
      name: space-limit-range
    spec:
      limits:
      - default:
          cpu: 1
        defaultRequest:
          cpu: 0.5
        type: Container
```

#### 5.2. Helm Chart Templates

```yaml template-helm.yaml
apiVersion: config.kiosk.sh/v1alpha1
kind: Template
metadata:
  name: redis
resources:
  helm:
    releaseName: redis
    chart:
      repository:
        name: redis
        repoUrl: https://kubernetes-charts.storage.googleapis.com
    values: |
      redisPort: 6379
```