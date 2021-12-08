---
title: "Kubernetes webhook 本地调试方法"
date: 2021-12-06T01:13:57Z
---

在 Kubernetes Webhook 开发过程中，调试是复杂场景下不可或缺的步骤。 但 Kubebuilder 文档中只给出了的远程部署运行步骤以及禁用 webhook 的方法，并没有给出本地调试的具体方法，只是了了数字带过：

> If you want to run the webhooks locally, you’ll have to generate certificates for serving the webhooks, and place them in the right directory (/tmp/k8s-webhook-server/serving-certs/tls.{crt,key}, by default).

> If you’re not running a local API server, you’ll also need to figure out how to proxy traffic from the remote cluster to your local webhook server. For this reason, we generally recommend disabling webhooks when doing your local code-run-test cycle, as we do below.

从以上内容中我们可以得到3点信息：
1. kube-apiserver 需要验证服务端证书验证 webhook 服务端是否合法
2. controller-mananger 默认从本地目录 `/tmp/k8s-webhook-server/serving-certs/tls.{crt,key}` 加载服务端证书
3. kube-apiserver 可以通过直接或代理连接到本地运行的 webhook 服务

接下来，我们讲解可直连的网络下的具体操作步骤。本文中我基于 kind 创建 kubernetes 集群， 因 kind 创建的集群是可以直接访问本机 IP 地址的。

## Webhook 注册方式

对于 MutatingWebhookConfiguration 和 ValidatingWebhookConfiguration 两种 webhook， 其 clientConfig 有两种配置方式：一是通过配置 url 地址的方式访问， 二是通过服务的方式进行配置。 配置的例子如下：

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
...
webhooks:
- name: my-webhook.example.com
  clientConfig:
    url: "https://my-webhook.example.com:9443/my-webhook-path"
  ...
```

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
...
webhooks:
- name: my-webhook.example.com
  clientConfig:
    caBundle: "Ci0tLS0tQk...<base64-encoded PEM bundle containing the CA that signed the webhook's serving certificate>...tLS0K"
    service:
      namespace: my-service-namespace
      name: my-service-name
      path: /my-path
      port: 1234
  ...
```

在本地调试时，我们可以使用本机 ip + 服务端口的方式配置 url， 并生成自签名证书 caBundle 来注册我们的本地服务器到 Kubernetes 集群。因此，我们还需要为我们的本地 IP 地址生成一个自签名证书。

```yaml
  clientConfig:
    url: "https://192.168.1.x:9443/my-webhook-path"
    caBundle: "Ci0tLS0tQk...<base64-encoded PEM bundle containing the CA that signed the webhook's serving certificate>...tLS0K"
```

## 生成自签名证书

下面步骤将生成一个自签名的 CA 证书以及服务端的 tls 证书。 需要注意的是 Go 1.15 版本开始废弃 CommonName，因此需要使用 SAN 证书。 


1. 根证书以及私钥创建

```bash
openssl genrsa -aes256 -out ca-key.key 4096
openssl req -new -x509 -days 365 -key ca-key.key -sha256 -subj "/CN=$HOST" -out ca.crt
```
修改 `$HOST` 和 `$IP` 环境变量，将其修改为本地调试机器的IP地址和主机名。当主机名不能被解析的时候，设置任何名称不会影响我们的调试。

2. 生成tls的密钥，并生成签名请求
```bash
openssl genrsa -out tls.key 4096

openssl req -new -sha256 \
    -key tls.key \
    -subj "/CN=$HOST" \
    -reqexts SAN \
    -config <(cat /etc/ssl/openssl.cnf \
        <(printf "\n[SAN]\nsubjectAltName=DNS:$HOST,IP:$IP")) \
    -out tls.csr
```

3. 签名tls证书

```bash
openssl x509 -req -days 365 \
    -in tls.csr -out tls.crt \
    -CA ca.crt -CAkey ca-key.key -CAcreateserial \
    -extensions SAN \
    -extfile <(cat /etc/ssl/openssl.cnf <(printf "[SAN]\nsubjectAltName=DNS:$HOST,IP:$IP"))
```

4. 查看服务器证书信息
```bash
openssl x509 -noout -text -in tls.crt
```

## 设置证书

1. 首先将 CA 证书导出 base64 字符串，并设置 clientConfig.caBundle

```bash
cat ca.crt | base64 |tr -d '\n'
```

2. 复制 ca.crt, tls.crt tls.key, 复制到 `/tmp/k8s-webhook-server/serving-certs/` 目录


## 最后
接下来，执行 `make install` 重新安装 webhook 并启动本地调试，不出意外情况下就可以愉快的调试了。

## 参考：

https://book.kubebuilder.io/cronjob-tutorial/running.html#running-webhooks-locally

https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/

https://kubernetes.io/blog/2019/03/21/a-guide-to-kubernetes-admission-controllers/