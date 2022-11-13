---
title: eGalax Touch USB 触摸屏驱动触摸事件优化
date: 2022-11-13T14:15:42Z
---


## 背景

最近在做的工控触摸屏项目，在硬件上使用到了 eGalax Touch 电阻触摸屏。该触屏的驱动支持常见 Linux 发行版，包括 Ubuntu Debian SuSE 等等。在软件层采了用基于 Electron 的客户端开发模式，运行在 Ubuntu 系统上。

该工程开发完成后，测试过程发现触屏上的交互行为非常不友好。因此，急需对触屏的体验进行优化。主要有以下几点问题：
1. 无法使用手指拖动滚屏，必须点击滚动条进行操作
2. 在屏幕上进行按压滑动会触发控件的拖动行为或者文字选择功能
3. 而在使用文字选择的时候几乎无法高亮想要选中的内容

## Level 1 触屏默认设置优化

eGalax Touch 自带了触屏配置工具，可以对触屏常见的可选项进行设置。在安装过 eGTouch_v2.5 即可在桌面上找到 **eGTouch Utility** 进行设置。在 **Setting** 标签页面下：
- 启用 "Enable Press and Hold" 允许长按事件。
- 禁用 “Enable Auto-Right Click“。 启用右键模拟后，会在长按后触发右键点击，进而导致长按丢失mouseup/touchend事件。
- 调整 "Press and Hold" ”Range“, 将其调整为最大值。

在硬件标签下，

- 适当降低灵敏度 Sensitivity

在我们的应用场景中，没有使用到右键，因此将右键模拟功能禁用。否则会在长按之后触发右键操作，导致长按失效。适当的调整触控范围和灵敏度可以进一步防止抖动。以上设置对浏览器上的拖拽，点击行为优化起到了一定的效果。然而，依然没有解决核心的页面滑动和文字选择。因此需要从应用层进行进一步优化。

## Level 2 Chromium 触屏模式

Electron 是基于 Chromium 内核开发的，因此，Electron 也是支持 Chromium 的触屏操作的。虽然选用的是电阻屏，只支持单点触摸。但是在触屏模式下，依然可以使用一些常见的触控操作的，比如向下滑动屏幕，向前/后翻页等功能。然而，虽然我们是触摸屏设备，但是 Chromium 却无法接收到任何 touch 事件或手势操作。

一番搜索过后，找到一篇《Fixing Chromium/Chrome Ignoring Touch Input on Linux》的文章。 上面文章中提到，Chromium 老版本中，会忽略触屏的 touch 事件。解决方法也相对简单.

1. 首先，在 Linux 下启动一个终端，输入以下命令：
```bash
xinput list
⎡ Virtual core pointer                          id=2    [master pointer  (3)]
⎜   ↳ Virtual core XTEST pointer                id=4    [slave  pointer  (2)]
⎜   ↳ Video Bus                                 id=6    [slave  pointer  (2)]
⎜   ↳ Power Button                              id=7    [slave  pointer  (2)]
⎜   ↳ eGalaxTouch Virtual Device for Single     id=17   [slave  pointer  (2)]
⎣ Virtual core keyboard                         id=3    [master keyboard (2)]
    ↳ Virtual core XTEST keyboard               id=5    [slave  keyboard (3)]
    ↳ Video Bus                                 id=18   [slave  keyboard (3)]
    ↳ Power Button                              id=19   [slave  keyboard (3)]
```
该命令会输出当前 X11 中找到的所有输入设备，从上面的列表中我们可以考到 *pointer* 和 *keboard* 两种设备类型。在 pointer 列表下的 *eGalaxTouch Virtual Device for Single* 就是我们所使用的触屏输入设备，其设备ID 为 17。

2. 然后，使用以下命令启动 Chromium:

```
chromium --touch-devices=17
```

`--touch-devices` 选项用于指定输入设备，并将该中断设备事件作为 touch 事件。注，该选项仅限于 "Xinput2"(X Server 1.8)版本以上的 Linux 桌面系统。

该选项同样适用于 Electron，使用该命令启动 Electron 之后，即可以使用简单的手势操作了。尤其在各种带有滚动条的列表中，可以轻易的滚动列表。

除此，以外可以设置 `--touch-selection-strategy` 优化文字选择操作。

然而，又出现了意料之外的情况。经过一系列的测试发现：

1. 长按情况下仅触发 `touchstart` 和 `touchend` 事件，无法监听到 `mousedown` 与 `mouseup`， 导致原来使用鼠标事件监听的操作无法完成。非触屏设备下无法使用。
2. 点击情况下，`touchstart` 和 `touchend` 与`mousedown` 与 `mouseup` 被依次触发。导致，同时注册以上事件时，注册的回调函数被执行两次。

经过一些列讨论后，前端一致认为该行为是驱动/操作系统问题，因此需要从底层解决。主要观点是，在没有 `--touch-devices` 参数时，Chromium 获取的都是鼠标事件，而没有touch 事件。通过`--touch-devices` 强制修改输入设备可能跟真正的触屏模式不同。而且在其他平板上可以通过属性判断是否为触屏设备，使用该方式无法判断是否为触屏设备。

因此，需要进一步的调研 Linux 系统中触屏的工作机制，以解决目前所遇到的问题。

## Level 3  Linux 输入设备与驱动

### Linux 输入设备基本知识

通过阅读 《Linux Input Subsystem userspace API》，可以理解一些输入设备的基本概念。
首先，Linux 输入事件被抽象为设备驱动和事件句柄，设备驱动负责与具体的硬件进行通讯，比如 USB 鼠标键盘，并提供 input 内核模块的 input 事件。事件句柄通过 evdev 事件接口与应用层进行通讯。在 Linux 下，一切设备都是文件，因此我们可以通过下面的命令查看输入设备：

```sh
ls  /dev/input

crw-r--r--   1 root     root      13,  64 Apr  1 10:49 event0
crw-r--r--   1 root     root      13,  65 Apr  1 10:50 event1
crw-r--r--   1 root     root      13,  66 Apr  1 10:50 event2
crw-r--r--   1 root     root      13,  67 Apr  1 10:50 event3
```

其次，在输入设备中，仅包含 *keyboard*，*mousedev*，*joydev* 等基本输入事件接口。并没有区分触屏或触控板。进一步阅读文档，我们可以看到输入设备可以产生不同的事件类型与事件代码。以鼠标为例，鼠标左键其产生的事件类型和代码应为 *EV_KEY*，*BTN_LEFT*；而移动鼠标则会发送，*EV_ABS*，*ABS_X*或*ABS_Y*。 而触控设备的事件类型与鼠标相同，皆为 *EV_KEY*，但是触屏设备的事件代码则为 *BTN_TOUCH*，而移动都使用 *EV_ABS*，*ABS_X* 或 *ABS_Y* 事件类型和代码。

为了进一步验证以上内容，我们可以使用 `evtest` 工具进行测试，输入 `evtest` 命令后，选择我们使用的 eGalaxTouch 设备：
```
evtest
/dev/input/event0:      Power Button
/dev/input/event7:      eGalax Inc. USB TouchController Stylus
/dev/input/event8:      eGalax Inc. USB TouchController
/dev/input/event14:     eGalaxTouch Virtual Device for Single
Select the device event number [0-14]: 
```
屏幕上会打印出该设备所支持的事件类型，以及事件代码。从该输出结果可以看出，我们目前的触摸板仅能支持鼠标点击事件。点击屏幕后，即可获取当前点击的坐标以及点击事件 *BTN_LEFT*

```
Input driver version is 1.0.1
Input device ID: bus 0x6 vendor 0xeef product 0x1 version 0x1
Input device name: "eGalaxTouch Virtual Device for Single"
Supported events:
  Event type 0 (EV_SYN)
  Event type 1 (EV_KEY)
    Event code 272 (BTN_LEFT)
    Event code 273 (BTN_RIGHT)
  Event type 3 (EV_ABS)
    Event code 0 (ABS_X)
      Value   1570
      Min        0
      Max     2047
    Event code 1 (ABS_Y)
      Value    713
      Min        0
      Max     2047
    Event code 3 (ABS_RX)
      Value      0
      Min        0
      Max     2047
    Event code 4 (ABS_RY)
      Value      0
      Min        0
      Max     2047

Event: time 1668136441.059613, -------------- SYN_REPORT ------------
Event: time 1668136441.125574, type 3 (EV_ABS), code 0 (ABS_X), value 393
Event: time 1668136441.125574, type 3 (EV_ABS), code 1 (ABS_Y), value 1488
Event: time 1668136441.125574, type 1 (EV_KEY), code 272 (BTN_LEFT), value 0
```
### eGalaxTouch 驱动

经过一番搜索和驱动文档阅读后，我们发现 eGalaxTouch 驱动可以通过配置文件修改事件代码类型。在 eGTouchL.ini 文件中有如下注释:

```
cat /etc/eGTouchL.ini

EventType                       [0]     AUTO_DETECT
BtnType                         [0] BTN_LEFT    [1] ABS_PRESSURE        [2] BTN_TOUCH
```
可见，我们仅需要将 `BtnType` 修改为2 即可将触屏的事件改为 `BTN_TOUCH`。然而，修改之后更严重的问题发生了。重启机器后，无论是 Ubuntu 系统桌面还是 Chromium 浏览器，屏幕上点击事件均不响应，仅可移动鼠标。而通过 `evtest` 测试，我们也的确得到了们想要的 `BTN_TOUCH` 事件，可以确认设备是正常工作的。

```bash
Event: time 1668137132.240446, -------------- SYN_REPORT ------------
Event: time 1668137132.326380, type 3 (EV_ABS), code 0 (ABS_X), value 537
Event: time 1668137132.326380, type 3 (EV_ABS), code 1 (ABS_Y), value 1327
Event: time 1668137132.326380, type 1 (EV_KEY), code 330 (BTN_TOUCH), value 0
```

## Level 4 X input 驱动与 Libinput

再次进入僵局后，我们将焦点转移到了 X Window。X Window 是 Linux 下的窗口管理系统，GNOME和KDE也都是以X Window系统为基础建构成的。基于 X 协议(X11，为X协议的第11版)的应用，由服务端和客户端两部分组成。X Server 通过 X 驱动连接输入设备。并将输入事件发送给客户端进行事件处理。在 Archlinux 的 wiki 我们获取到 TouchScreen 的 X11 驱动通常是即插即用的。常见的驱动包括：f86-input-evdev、xf86-input-libinput、xf86-input-elographics 等等。X 驱动与设备的映射关系则是通过 `xorg.conf` 进行配置。以 eGalaxTouch 安装包自带的驱动配置为例：

```
# /usr/share/X11/xorg.conf.d/52-egalax-virtual-libinput.conf

Section "InputClass"
	Identifier "eGalax virtual class"
	MatchProduct "eGalaxTouch Virtual Device"
	MatchDevicePath "/dev/input/event*"
	Driver "libinput"
EndSection
```
当前配置指定了，当匹配到 eGalaxTouch 的设备时，使用 libinput 驱动，下图更好的说明了内核，X Server，libinput 之间的关系。

![Libinput](https://wayland.freedesktop.org/libinput/doc/latest/_images/graphviz-27435b39bb1f5a51560bcd647da4891135378214.png)

另外，在X11 input 设备驱动中有定义，设备的类型。如：XI_MOUSE、 XI_TABLET、 XI_KEYBOARD、 XI_TOUCHSCREEN、 XI_TOUCHPAD 等。那么使用 libinput 驱动时是如何进步地确定设备类型呢？在进一步的进行深入研究驱动层前，我们先看一下现在设备的属性和 libinput 事件。

### Libinput 调试

在上一步中我们已经得知 eGalaxTouch 在内核驱动层是正常工作的。Libinput 作为内核与 X Server的桥梁，我们需要进一步的分析 Libinput 是否正常工作。

首先，我们需要安装 `libinput-tools` 调试工具。

```bash
apt install libinput-tools
```

将 eGalaxTouch 模式，首先改为 *BTN_LEFT*。 然后监听 libinput 事件。

```
# sudo libinput debug-events
 event4   POINTER_MOTION_ABSOLUTE +26.65s        56.69/ 90.48
-event4   POINTER_BUTTON   +26.67s      BTN_LEFT (272) pressed, seat count: 1
 event4   POINTER_BUTTON   +26.86s      BTN_LEFT (272) released, seat count: 0
```

再次将 eGalaxTouch 模式修改为 *BTN_TOUCH*。 然后监听 libinput 事件，虽然我们修改了驱动层的事件代码，但是当前上报的依然是POINTER_MOTION_ABSOLUTE 事件。并且此时 *BTN_LEFT* 的事件已经不再触发了。

```
# sudo libinput debug-events
event4   POINTER_MOTION_ABSOLUTE +26.65s        56.69/ 90.48
```


在找到了另外一块触摸屏产品后，对两块屏幕的属性和事件进行了对比。以下是 TSC2007 触屏的设备属性及触屏事件。最直接的区别就是 POINTER_* 事件被替换成了 TOUCH_* 事件。而且在能力(Capabilities) 我们也可以看出，eGalaxTouch 中为 pointer， 而 TSC2007 则为 touch。

``` 触屏事件
# sudo libinput debug-events
 event1   TOUCH_DOWN       +36.40s      -1 (0) 24.47/76.74 (196.00/983.00mm)
 event1   TOUCH_MOTION     +36.40s      -1 (0) 24.47/76.74 (196.00/983.00mm)
 event1   TOUCH_FRAME      +36.40s
 event1   TOUCH_UP         +36.41s

# sudo libinput list-devices
Device:           TSC2007 Touchscreen
Kernel:           /dev/input/event1
Group:            3
Seat:             seat0, default
Capabilities:     touch
Tap-to-click:     n/a
Tap-and-drag:     n/a
Tap drag lock:    n/a
Left-handed:      n/a
Nat.scrolling:    n/a
Middle emulation: n/a
Calibration:      identity matrix
Scroll methods:   none
Click methods:    none
Disable-w-typing: n/a
Accel profiles:   n/a
Rotation:         n/a
```

### 设备类型

回顾上一节内容，input 设备在内核层对输入设备是没有触摸板、触屏、鼠标等设备类型定义的。对于鼠标和触摸板，均使用 *EV_ABS* 事件类型，仅按键 *EV_KEY* 事件代码不同。而 Libinput 层对设备和事件进行了封装，将不同的设备的事件转换为 *TOUCH_** 或 *POINTER_** 事件，推送给 X input 驱动。

那么对于 Libinput 怎么判断判断设备的类型呢？我们进一步查阅官方文档，对设备类型的定义。

简单的总结就是 Libinput 不对上层应用暴露设备类型，仅仅提供其所具有的 **能力（Capabilities）** 和 **特性（features）**。例如，触摸板对于 Libinput 是一个具有坐标 (pointer)和手势(gestures)的设备，并且具有一系列的选项配置，如Tap-to-Click。

但是， Libinput 会通过 Udev 的标签进行判断设备的属性，进而确定设备的类型，并将设备产生的原生事件进行转换。规则很简单，如一个设备的 udev 属性中带有 ID_INPUT_KEYBOARD， Libinput 就会将其看作是键盘。同理如果带有，ID_INPUT_TOUCHSCREEN 怎会被看作是触摸屏。

### udev 修改设备属性

在此不再进一步展开 udev 相关内容。仅给出方法，udev 可以通过规则文件监听设备事件，进而执行任务或者修改属性等操作。我们利用这一特性，修改 eGalaxTouch 的属性。按以下规则修改 `/etc/udev/rules.d/` 下的规则文件。

```
# /etc/udev/rules.d/99-egalax-udev.rules

# 原有内容，具体作用,猜测是移除物理设备 eGalax Inc. USB TouchController 事件源，进使用虚拟设备。
ACTION=="add", ENV{ID_VENDOR_ID}=="0eef", ENV{ID_INPUT}=""

# 新加内容，匹配(==) eGalaxTouch 属性,并设置 ID_INPUT_TOUCHSCREEN = 1
ACTION=="add|Change",SUBSYSTEM=="input", ENV{NAME}=="eGalaxTouch Virtual Device for Single", ENV{ID_INPUT_TOUCHSCREEN}="1"

```

再次重启后，Ubuntu 桌面系统终于恢复了正常所有，所有点击均可正常使用，并且可以使用简单的手势操作。

## Level 2.1 Chromium Touch 事件

经过以上设置后，Chromium 已经完全进入了触摸模式，已经不需要再设置 `--touch-devices` 参数。但是 `touchstart` 和 `mousedown` 依次触发的问题依然没有解决。无耐之下，进一步查看浏览器对事件的处理。然而，这确实是浏览器的默认行为，而解决方案也比较简单。 在 `touchstart`和其他 touch 事件后调用 `preventDefault()` 或 `stopPropagation()` 即可阻止 `mousedown` 的触发。具体的请看MDN 文档。

## 最后
完了，没有总结

## 参考：
- https://www.eeti.com/drivers_Linux.html
- https://wiki.archlinux.org/title/Touchscreen
- https://www.kernel.org/doc/html/v4.12/input/input.html#architecture
- https://www.x.org/wiki/Development/Documentation/XorgInputHOWTO/
- https://wayland.freedesktop.org/libinput/doc/latest/what-is-libinput.html
- https://yuenhoe.com/blog/2015/06/fixing-chromiumchrome-ignoring-touch-input-on-linux/
- https://developer.mozilla.org/en-US/docs/Web/API/Touch_events#additional_tips