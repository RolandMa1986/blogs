
---
title: '深入浅出 Kubernetes 项目网关与应用路由'
date: 2021-11-19T13:23:25Z
draft: true
---

# 定制化 Ubuntu 镜像的几种方式

## 从零开始

```
sudo debootstrap \
   --arch=amd64 \
   --variant=minbase \
   bionic \
   $HOME/live-ubuntu-from-scratch/chroot \
   http://mirrors.ustc.edu.cn/ubuntu/
```

```
cat <<EOF > /etc/apt/sources.list
deb http://mirrors.ustc.edu.cn/ubuntu/ bionic main restricted universe multiverse 
deb-src http://mirrors.ustc.edu.cn/ubuntu/ bionic main restricted universe multiverse 

deb http://mirrors.ustc.edu.cn/ubuntu/ bionic-security main restricted universe multiverse 
deb-src http://mirrors.ustc.edu.cn/ubuntu/ bionic-security main restricted universe multiverse 

deb http://mirrors.ustc.edu.cn/ubuntu/ bionic-updates main restricted universe multiverse 
deb-src http://mirrors.ustc.edu.cn/ubuntu/ bionic-updates main restricted universe multiverse    
EOF
```

参考：
https://github.com/mvallim/live-custom-ubuntu-from-scratch


## Cubic 
Cubic(Custom Ubuntu ISO Creator) 是一个向导式的 Ubuntu Live ISO 创建工具。
