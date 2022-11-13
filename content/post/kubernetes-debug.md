
---
title: "Hello world"
date: 2020-10-21T20:43:47+08:00
draft: true
categories: ["others"]
---
复制 dlv
ubuntu@host:~$ scp -i ~/.minikube/machines/minikube/id_rsa  ~/go/bin/dlv docker@172.17.0.3:

复制 apiserver

ubuntu@host:~$ scp -i ~/.minikube/machines/minikube/id_rsa /home/roland/go/src/k8s.io/kubernetes/apiserver   docker@172.17.0.3:

查看kube 命令：

docker@minikube:~$ sudo cat /etc/kubernetes/manifests/kube-apiserver.yaml

运行 apiserver

sudo ./dlv exec ./apiserver --headless --listen=:2345 --log --api-version=2 -- --advertise-address=172.17.0.3 --allow-privileged=true --authorization-mode=Node,RBAC --client-ca-file=/var/lib/minikube/certs/ca.crt --enable-admission-plugins=NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota --enable-bootstrap-token-auth=true --etcd-cafile=/var/lib/minikube/certs/etcd/ca.crt --etcd-certfile=/var/lib/minikube/certs/apiserver-etcd-client.crt --etcd-keyfile=/var/lib/minikube/certs/apiserver-etcd-client.key --etcd-servers=https://127.0.0.1:2379 --insecure-port=0 --kubelet-client-certificate=/var/lib/minikube/certs/apiserver-kubelet-client.crt --kubelet-client-key=/var/lib/minikube/certs/apiserver-kubelet-client.key --kubelet-preferred-address-types=InternalIP,ExternalIP,Hostname --proxy-client-cert-file=/var/lib/minikube/certs/front-proxy-client.crt --proxy-client-key-file=/var/lib/minikube/certs/front-proxy-client.key --requestheader-allowed-names=front-proxy-client --requestheader-client-ca-file=/var/lib/minikube/certs/front-proxy-ca.crt --requestheader-extra-headers-prefix=X-Remote-Extra- --requestheader-group-headers=X-Remote-Group --requestheader-username-headers=X-Remote-User --secure-port=8443 --service-account-key-file=/var/lib/minikube/certs/sa.pub --service-cluster-ip-range=10.96.0.0/12 --tls-cert-file=/var/lib/minikube/certs/apiserver.crt --tls-private-key-file=/var/lib/minikube/certs/apiserver.key


