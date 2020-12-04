---
title: "Kubernetes RBAC - 证书认证"
date: 2020-12-04T00:37:34Z
tag: 
- Kubernetes
---

## X509 客户证书

一般 Kubernetes 集群都会启用基于客户证书的用户认证模式。 API Server 通过 `--client-ca-file=SOMEFILE` 选项配置 CA 证书用于客户证书认证。如果提供的证书被验证通过，则 subject 中的公共名称（Common Name）就被作为请求的用户名。 通常可以使用两种方式签发：

- 通过 CertificateSigningRequest 资源类型允许客户申请签名 X.509 证书
- 使用外部证书服务颁发证书，如 openssl

## 通过 CSR API 签发证书步骤：

1. 生成用户私钥：

```bash
openssl genrsa -out user2.key 2048
```

2. 使用私钥生成证书签名请求 CSR:

```bash
openssl req -new -key user2.key -out user2.csr -subj "/CN=user2/O=group1/O=group2"
```

> 如果遇到错误信息 `Can't load /home/ubuntu/.rnd into RNG` 需先执行：
```bash
openssl rand -writerand .rnd
```

3. 提交 CSR 请求：

```bash
kubectl apply -f - <<EOF
apiVersion: certificates.k8s.io/v1beta1
kind: CertificateSigningRequest
metadata:
  name: user2
spec:
  request: $(cat user2.csr | base64 | tr -d '\n')
  usages: ['digital signature', 'key encipherment',
    'client auth']
EOF
```

查看 CSR 请求：

```bash
kubectl get csr
NAME    AGE   SIGNERNAME                     REQUESTOR          CONDITION
user2   75s   kubernetes.io/legacy-unknown   kubernetes-admin   Pending
```

4. 批准或拒绝 CSR：

```bash
kubectl certificate approve user2
kubectl certificate deny user2
```

5. 获取签发的证书并导出

```bash
kubectl get csr user2 -o jsonpath='{.status.certificate}' | \
base64 --decode > user2.crt
```

6. 添加到 KubeConfig
```bash
kubectl config set-credentials user2 --client-key=./user2.key --client-certificate=user2.crt --embed-certs=true

kubectl config set-context user2 --cluster=kubernetes --user=user2

kubectl config use-context user2
```

## 手动签发证书

1. 生成用户私钥:

```bash
openssl genrsa -out user1.key 2048
```

2. 使用私钥生成证书签名请求 CSR:

```bash
openssl req -new -key user1.key -out user1.csr -subj "/CN=user1/O=group1/O=group2"
```

3. 使用CA证书签发用户证书:

```bash
openssl x509 -req -in user1.csr \
    -CA /etc/kubernetes/pki/ca.crt \
    -CAkey /etc/kubernetes/pki/ca.key \
    -set_serial 101 -extensions client -days 365 -outform PEM -out user1.crt
```
> /etc/kubernetes/pki/ca.crt 为使用 kubeadm 安装 kubernetes 后默认生成的 CA 证书。

4. 添加到 KubeConfig 步骤同上

## 参考：

https://kubernetes.io/zh/docs/reference/access-authn-authz/certificate-signing-requests/#nornal-user

https://kubernetes.io/zh/docs/reference/access-authn-authz/authentication/