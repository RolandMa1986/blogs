---
title: "使用 apt 命令获取 Ubuntu 安装包源码"
date: 2020-12-14T15:03:27Z
tag: 
- ubuntu
- systemd
- linux
---

由于最近需要研究一个关于 Ubuntu 上 systemd 的问题，因此需要下载 systemd 的源代码。 一般源码下载可以在官网或者Github中下载，如 systemd。但是由于 linux 各个分发版本可能会给它们打上自己的补丁，因此我们需要从 Ubuntu 下载其源码包。在 Debian 或 Ubuntu 的系统中源码是文件版的软件发行包，因此我们可以使用 apt-get 或者 apt 命令下载其源码 (DEB 文件包)。

<!--more-->

## Ubuntu 下载步骤

### Step 1 启用源码仓库

Ubuntu 的源码仓库默认为禁用状态。因此在下载安装包之前，需要先启用源码仓库。首先编辑 APT 源文件 /etc/apt/sources.list：

```bash
$ sudo vi /etc/apt/sources.list
```

以 Ubuntu 18.04 为例，取消所有以 deb-src 为开头的源路径。

```
deb-src http://cn.archive.ubuntu.com/ubuntu bionic main restricted
deb-src http://cn.archive.ubuntu.com/ubuntu bionic-updates main restricted
deb-src http://cn.archive.ubuntu.com/ubuntu bionic universe
...
```
其中：
1. `deb-src` 表明其为deb的源码仓库 
2. `http://cn.archive.ubuntu.com/ubuntu` 是获取 index 和 deb 文件的URL
3. `bionic` 为 Ubuntu 18.04 tls 发行版代号
4. `main, universe` 不同组件的仓库

我们也可以添加第三方的源已经代码源，以 docker-ce 为例在源中增加以下两行即可:

```
deb [arch=amd64] https://download.docker.com/linux/ubuntu bionic stable
deb-src [arch=amd64] https://download.docker.com/linux/ubuntu bionic stable
```

### Step 2 更新 index

修改过源文件之后需要执行 `apt update` 或者 `apt-get update` 命令更新包索引文件。

### Step 3 安装 dpkg-dev 包

下载源码依赖于 dpkg-dev 进行解包。
```
sudo apt-get install dpkg-dev
```
如果未安装 dpkg-dev 执行下载时会遇到以下错误：

> sh: 1: dpkg-source: not found
> W: Download is performed unsandboxed as root as file 'systemd_237-3ubuntu10.43.dsc' couldn't be accessed by user '_apt'. - pkgAcquire::Run (13: Permission denied)
> E: Unpack command 'dpkg-source --no-check -x systemd_237-3ubuntu10.43.dsc' failed.
> N: Check if the 'dpkg-dev' package is installed.

### Step 4 下载 Ubuntu 包源码

接下来我们就可以使用 apt 命令来下载源码了。我们以 systemd 为例演示下载步骤:

```bash
# sudo apt-get source {pkg name} 
$ sudo apt-get source systemd
```

至此 systemd 已经被下载并解压到当前工作目录。

## 理解源码结构

首先使用 ls 命令查看当前目录：

```bash
ls -l
total 7000
drwxr-xr-x 28 root root    4096 Dec 15 01:44 systemd-237
-rw-r--r--  1 root root  277780 Oct 26 15:13 systemd_237-3ubuntu10.43.debian.tar.xz
-rw-r--r--  1 root root    5187 Oct 26 15:13 systemd_237-3ubuntu10.43.dsc
-rw-r--r--  1 root root 6871350 Jan 31  2018 systemd_237.orig.tar.gz
```
在这个当前文件夹下：

- `systemd_237.orig.tar.gz` 以 .orig.tar.gz 结尾的文件为上游源码的压缩包。
- `systemd_237-3ubuntu10.43.dsc` 为当前包的描述文件，包括依赖项，签名等信息。
- `systemd_237-3ubuntu10.43.debian.tar.xz` debian/ubuntu 发行版对上游源码包修改的patch文件压缩包。
- `systemd-237` 解压后的源代码目录

下载源码时，我们可以使用 --download-only 选项，来跳过解压过程。下载完成后，可以使用 dpkg-source 命令进行解压.

```bash
$ sudo apt-get --download-only source {pkg}
$ dpkg-source -x /path/to/pkg.dsc
```

## 编译

使用 apt-get build-dep 命令可以为我们自动安装编译 systemd 所需要的工具包：

```
sudo apt-get -y build-dep systemd
```

最后，我们可以自行打包：

```
$ debuild
```

至此，我们已经得到了源码，在此基础之上编译了自己deb包。

参考：
https://www.cyberciti.biz/faq/how-to-get-source-code-of-package-using-the-apt-command-on-debian-or-ubuntu/