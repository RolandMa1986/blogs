---
title: "基于 Minimal Linux Live 构建自动刷镜像U盘"
date: 2023-05-01T14:16:47+08:00
draft: true
---

最近在生产一款采用 Linux 系统的 X86 架构工控机。工控机生产过程中，通常采用 Etcher 或其它烧录工具配合1拖N硬盘对拷机进行系统镜像烧录。然而，在客户现场升级或返厂维修时，如采用上述方法则需要将工控机拆机后烧录，过程较为繁琐。因此，急需一款类似 Ghost 的工具实现自动化刷机的工具及自启动U盘。在探索过程中，初步确定了基于 Minimal Linux Live，自定义 shell 脚本及系统镜像的自启动U盘刷镜像方案。 

Minimal Linux Live(以下简称MLL)是一个非常小巧的教育性 Linux 发行版，它是通过一系列自动化 Shell 脚本从头构建的。MLL提供了一个包含 Linux 内核、GNU C 库和 Busybox 等工具的核心环境。

<!--more-->

## MLL 的构建与使用

以下构建步骤基于 Ubuntu 18.04（20.04）系统。虽然 MLL 构建的iso 文件仅 10MB 大小，但是需要至少 2GB 的可以用硬盘空间，用于构建内核及其他依赖包。

1. 从 Github 中下载最新的 Release 版本 [15-Dec-2019](https://github.com/ivandavidov/minimal/releases)到本地或克隆源代码：
   ```bash
   $ git clone https://github.com/ivandavidov/minimal.git
   $ git checkout 15-Dec-2019
   ```
2. 安装构建依赖库，如 GCC、make 等，在Ubuntu 系统下直接使用以下命令安装：
   ```bash
   $ sudo apt install wget make gawk gcc bc bison flex xorriso libelf-dev libssl-dev
   ```
3.  执行 src 目录下的构建脚本 `build_minimal_linux_live.sh`,构建过程将持续20~30左右。构建完成后将会得到 `minimal_linux_live.iso` 与 `mll_image.tgz` 文件。
   ```bash
   $ src/build_minimal_linux_live.sh
   ```

> 注意，使用在使用 Ubuntu 默认的 gcc 时，构建 glibc 会报错。修复方法如下：
 ```bash
#文件位置： src/04_build_glibc.sh 
echo "Configuring glibc."

$GLIBC_SRC/configure \
  --prefix= \
  --with-headers=$KERNEL_INSTALLED/include \
  --without-gd \
  --without-selinux \
  --disable-werror \
  --enable-cet \ # 增加该行
  CFLAGS="$CFLAGS 
```

接下来我们可以使用虚拟机加载 `minimal_linux_live.iso` ISO 镜像文件用以启动 MLL或将 MLL 以 ISO 格式写入 U盘，作为启动盘。最简单的方式是使用`dd`命令,如下所示：

```bash
# 注：/dev/xxx为U盘的路径
dd if=minimal_linux_live.iso of=/dev/xxx
```

这样U盘将被识别为一个可启动设备，在BIOS中设置U盘为启动设备后即可进入MLL。

## MLL 启动过程

MLL 使用 SYSLINUX 作为boot loader。SYSLINUX 是一个 兼容 CD,U盘和PXE等多种设备的启动引导器，同时支持 BIOS 及 UEFI。下文中我们会以 BIOS 启动过程为例，分析 MLL 的其启动过程。在分析启动过程前，我们首先熟悉一下 MLL 镜像文件的结构。

### ISO 镜像结构

当我们在配置文件 `.config` 中将 `FIRMWARE_TYPE` 属性设置为 `bios` 时，打包完成的 iso 镜像结构如下：

```
minimal_linux_live.iso
├── boot/
│   ├── kernel.xz
│   ├── rootfs.xz
│   └── syslinux/
├── EFI/
└── minimal/
```

- **boot/** 文件夹下包含 `BIOS` 启动过程中需要的所有文件。在该文件下可以找到 Linux 内核、initramfs及 SYSLINUX 启动引导。
- **boot/kernel.xz** Linux 内核，内核初始化后会检测硬件信息并加载驱动，然后将控制权移交到 `initramfs`
- **boot/rootfs.xz** 即 `initramfs` 文件系统。initramfs 在内核启动的早期提供用一个户态环境，用于完成在内核启动阶段不易完成的工作。内核启动后会自动解压该文件到内存文件系统中。实际的控制权移交工作由 `/init` shell 脚本完成。
- **boot/syslinux/** 文件夹包含了 ISOLINUX 引导器的可执行文件和配置文件，它是 SYSLINUX 的子项目。
- **minimal/** 文件夹包含了 MLL 的附加包，大多数MLL的工具由附加包提供，如 openjdk，DHCP等。

### 启动流程



**init 用户空间启动过程**
```
# System initialization sequence:
#
 /init (this file)
  |
  +--(1) /etc/01_prepare.sh
  |
  +--(2) /etc/02_overlay.sh
          |
          +-- /etc/03_init.sh
               |
               +-- /sbin/init
                    |
                    +--(1) /etc/04_bootscript.sh
                    |       |
                    |       +-- /etc/autorun/* (all scripts)
                    |
                    +--(2) /bin/sh (Alt + F1, main console)
                    |

```


## 参考
https://wiki.archlinux.org/title/syslinux

https://github.com/ivandavidov/minimal

https://ivandavidov.github.io/minimal/#home