---
title: "基于Ubuntu18.04的云原生应用开发环境搭建VSCode&Golang"
date: 2020-09-22T01:42:38Z
tag: 
- golang
- vscode
- ubuntu
---

<!--more-->

## 前提条件

已安装Ubuntu18.04桌面版

## Golang 安装

1. 下载GO安装包

国内用户可以使用Google的中国区镜像站点https://golang.google.cn下载最新的golang安装包。这里我们选用基于linux的64位发行版。
```bash
wget https://golang.google.cn/dl/go1.14.9.linux-amd64.tar.gz
```

2. 解压tar包到用户目录

```bash
sudo tar -C /usr/local -xzf go1.14.7.linux-amd64.tar.gz
```
3. 设置Path环境变量,并使其生效

```bash
# Go 
echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.profile
source ~/.profile
```
4 测试Go是否安装成功

```bash
go version
```

5. 设置 GoProxy
使用go get安装go依赖包依赖于golang.org网络，国内用户可以使用goproxy设置镜像加速。对于Go1.13以上版本执行以下命令:

```bash
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
```
> ref https://goproxy.cn/

6. GOPATH

GOPATH为可选配置，根据个人需求进行配置。

## VScode 开发环境 

1. 下载 vscode

进入[Vscode官网](https://code.visualstudio.com/)，并选择deb格式安装包：
https://code.visualstudio.com/

2. 安装Go plugin

安装成功后选择扩展按钮或Ctrl+Shift+X进入扩展面板，输入`Go`搜索并安装Go插件。

3. 激活Go扩展

打开任意包含Go源码的目录就可以激活Go插件,或按Ctrl+Shift+P 打开命令，输入`Go: Install/Update Tools`安装Go需要的插件。

4. 设置最大文件检查数量
打开包含大量文件的工程时需要设置linux的最大文件监控数量。

``` bash
$ sudo nano /etc/sysctl.conf
#增加到最后一行
fs.inotify.max_user_watches=524288

$sudo sysctl -p
```
