---
title: "Kubernetes Issue 归类指南"
date: 2020-11-06T08:42:33Z
---


> Issue 管理是软件工程中重要的一个环节。项目管理者可以通过 issue 对项目做持续的追踪，复审，把握整体的质量和进度。同时及时发现风险和问题。在开源软管理过程中，issue 的追踪变得更为棘手，来自全世界开发者的 issue 会源源不断的进入 issue 管理系统中。因此更需要一个高效的管理流程及工具。

<!--more-->

> 本文原文源自 Kubernets 社区的 issue 归类指南 [Issue Triage Guidelines](https://www.kubernetes.dev/docs/guide/issue-triage/#what-is-triaging)， 以此窥探 Kubernets 社区是如何高效的管理、追踪 issue。


# Issue 归类：入门指南

## 适用范围

本指南作为 Kubernets issue 归类依据的主要文档。同时鼓励 SIG 和项目使用本指南作为基础，并根据项目需求做定制化。
**注意:** 本指南只适用于 Kubernetes 仓库。对于其他 Kubernetes 相关的 Github 仓储暂无强制要求。

## 什么是 Issue 归类？

Issue 归类是兴趣组（SIG）接收并复核 Github issue 或请求的处理流程，使它们能够被及时处理(本 GIS 的成员或者分配到其他  SIG)。归类包含对 issue 和 pull request 的优先级/紧急性，SIG 所有者，issue 类型(bug, feature, etc.)等属性的分类。

- SIG 负责 issue 和 pull request 的处理

归类可以在定期会议上开展，也可以是异步，持续的过程。很多 Kubernetes SIG 和项目已经形成了自己的模式。

> 例如，可以指派专人每天或者每周进行一次归类。

## 归类有哪些益处？

- 加速 issue 处理、管理
- 通过缩短响应时间保持贡献者的参与度
- 防止工作被无限拖延
- 以一种中立的流程作为标准，对“特别请求“以及个案问题进行归类
- 加强透明性，吸引更多人参与讨论，更多合作与周密的决策
- 帮助提高优先级管理，协商，决策能力等技术人员急缺的技能
- 巩固 SIG 社区和社区文化

对于喜欢项目管理和迭代流程的人更倾向于使用归类，因为它可以使 SIG 工作保持稳定、连续， 并且这些工作是依据用户反馈及其价值进行评估的，并列为高优先级的。

> issue 是一个重要的沟通途径，无论是社区用户、贡献者或是SIG团队成员都可以通过 issue 进行交流。 它即是获取用户声音的一个途径，也可作为团队内部协作的一种方式。因此有效的、合理的利用 issue 可以提高项目的协作、提高生产效率。

> 制定一个 issue 归类准则尤为重要，可以防止项目由于不断的加入“特别请求”而导致项目需求分化。项目的架构设计中应对扩展友好，让用户对个案问题进行扩展。同时保证核心架构稳定，将有限的精力集中在核心范围。

## 如何归类：手把手教学

本节目标是带你逐步的了解一个标准的归类流程。

> 原文中先讲的工具与技巧。但是个人认为建立流程和文化比工具使用更加重要，因此我们只关注流程。

### 第一步：审查新创建的 Issue

Kubernetes 使用 Github 进行 [issue](https://github.com/kubernetes/kubernetes/issues) 追踪。新创建的未归类 issue 是不带任何标签的。 SIG 管理者应该指派至少一名 SIG 成员作为 issue 的第一联系人。

标签是归类的主要工具，完整的标签列表可以查看 Github [Labels](https://github.com/kubernetes/kubernetes/labels)。 

#### 查询

Github 允许你过滤出不同类型的 issue 和 pull request， 它能帮助你找到需要归类的 issue。下表中包含了一些预定义的搜索条件用以方便查询。

搜索 | 搜索条件
---|---
[created-asc](https://github.com/kubernetes/kubernetes/issues?q=is%3Aissue+is%3Aopen+sort%3Acreated-asc) | 按创建时间由早到晚的未归类 issue 
[needs-sig](https://github.com/kubernetes/kubernetes/issues?q=is%3Aissue+is%3Aopen+label%3Aneeds-sig) | 需要被分配到 SIG 的 issue
[is:open is:issue](https://github.com/kubernetes/kubernetes/issues?q=is%3Aopen+is%3Aissue) | 最新的 issue
[comments-desc](https://github.com/kubernetes/kubernetes/issues?q=is%3Aissue+is%3Aopen+sort%3Acomments-desc) | 最受关注的未归类 issue， 按照评论数量从多到少排序 
[comments-asc](https://github.com/kubernetes/kubernetes/issues?q=is%3Aissue+is%3Aopen+sort%3Acomments-asc) | 需要更多关注的 issue 按照评论数量排序从少到多排序

我们建议从年代最久远的未归类 issue 或者 pull request 开始您的归类。

> Github issue 是开源软件中最流行的一种管理方式，其功能非常简洁，当然你可以使用一些第三方插件对它们更有效的管理。其他商用软件如 Jira 和 Azure DevOps 等功能都非常强大，同时管理流程也更为复杂。issue 管理工具的选择需根据自身团队文化，管理流程做出合适的选择，当软件无法满足管理需求时再做迁移即可。

### 第二步：根据 Issue 类型归类

使用链接中的[标签](https://github.com/kubernetes/kubernetes/labels?utf8=%E2%9C%93&q=triage%2F+kind%2Fsupport)能帮助你定位到哪些 issue 是可以快速关闭的。 负责归类的工程师可以添加合适的标签。

根据您的权限可以直接关闭或者使用标签命令备注那些被确定为帮助请求，重复，不能重现，或者那些缺少足够信息的 bug。

#### 帮助请求

有些人会错误的使用 Github issue 填写帮助请求。他们常常请求帮助他们做某些 Kubernetes 相关配置。对于这些 issue， 应该指导用户使用 [Slack](https://www.kubernetes.dev/docs/guide/issue-triage/#support-requests-channels) 支持频道。然后打上 `close` 与 `kind/support` 标签，引导用户使用我们的支持方式。

更详细的说明请参考备注中的[帮助请求](#)。

#### 废弃的或者发错位置的 issue

直接关闭或添加备注。

#### 需要更多信息

`triage/needs-information` 标签表示这个 issue 需要更多的信息才能继续在上面工作，添加备注或关闭。

> 过滤无效信息或缺少必要信息的 issue 是整个环节的第一步，它可以帮助其他贡献者或团队成员集中的精力于当前的工作。不会因 issue 的泛滥导致无从下手，或者大量精力花费再理解这些 issue 上。

#### Bugs

首先，尝试重现这个问题来验证它是不是一个bug。

如果你能重现它：

- 定义它的[优先级](https://www.kubernetes.dev/docs/guide/issue-triage/##step-three-define-priority)
- 做一次重复问题搜索，来确定这个问题是否已经被上报过。如果找到了重复项，关联原始的 issue 告知报告者， 并将它关闭。

如果你不能重现它：

- 联系上报者告并知你所做的尝试
- 如果双方都同意这个问题不可复现则将它关闭

如果你需要更多的信息才能继续工作：

- 通过标记 `lifecycle/needs-information` 标签让上报者知道。

在以上情况中，如果你在20天内得不到任何回应，则应该留下合适的评论并关闭这个 issue。如果你有权限关闭其他人的 issue，首先将它分配给自己 `/assign`，然后关闭它 `/close`。如果没有，请将你的调查结果写在评论中。

#### Help Wanted/Good First Issues

为了显式的标记那些对新贡献者友好的 issue， 我们使用 [help wanted](https://github.com/kubernetes/kubernetes/issues?q=is%3Aopen+is%3Aissue+label%3A%22help+wanted%22) 和 [good first issue](https://github.com/kubernetes/kubernetes/issues?q=is%3Aopen+is%3Aissue+label%3A%22good+first+issue%22) 标签。 

- 审阅我们的[准则](https://www.kubernetes.dev/docs/guide/help-wanted)学习如何使用他们
- 如果 issue 符合这些准则，你可以通过 `/help` 命令添加 `help wanted` 标签和通过 `/good-first-issue` 添加 `good first issue` 标签。请注意，添加 `good first issue` 的同时也会自动添加 `help wanted` 标签。
- 如果一个 issue 有上述标签，但是不符合准则，请为 issue 添加更多详细信息或者使用 `/remove-help` 或 `/remove-good-first-issue` 命令移除标签。

#### Kind 标签

通常 `kind` 标签由 issue 的提交者标记。 负责归类的工程师应更正这些错误的标签（如将帮助请求标记为bug）。复核是一个很好的纠错途径。我们的 issue 模板用于指导用户选择正确的分类。

### 第三步： 定义优先级

我们使用 Github 标签排列优先级。如果一个 issue 没有 `priority` 标签， 意味着它还未被审阅。

我们的目标要保证优先级在整个项目中是一致的。然而，如果你确信你发现了一个错误的 issue 优先级，请在评论中留下你的建议，我们会对此重新评估。

Priority 标签 | 含义 | 案例
---|---|---
priority/critical-urgent | 团队负责人应保证这些 issue 正在被处理-请放下手上其他工作。这些 issue 应在下个版本前被修复。 | 例如核心模块明显影响用户的bug，Broken build ，严重的安全问题。
priority/important-soon | 当前或很快就必须具备的和工作的功能。非常重要，但是不会是当前版本发布的障碍。 | [XXXX]
priority/important-longterm | 对于长期很重要，但是当前可以不具备或者需要多个版本才能完成。不会阻碍当前的版本发布。 | [XXXX]
priority/backlog | 一致认为这是应该拥有的功能，但是短期内没有足够的人手工作在上面。同时欢迎社区贡献，但可能等待较长时间才能得到 Reviewer 响应。 | [XXXX]
priority/awaiting-more-evidence | 可能是有用的，但是当前的信息还不足以支持实际投入 |  [XXXX]

### 第四步：找到并设置合适的 SIG，负责该 issue

Kubernetes 组件被划分为很多的工作组（SIG）。 

- 例如，使用/sig network 命令会增加 `sig/network` 标签。
- 名字含有多个英语单词的 SIG 使用“-”分割。 例如， `/sig cluster-lifecycle`。
- 请注意这些命令必须每条一行，并且再评论内容之前。
- 如果你不确定该 issue 由谁负责，仅需要使用 SIG 标签。
- 如果你觉得有必要提醒某个团队，可以使用@符号提醒，格式如下：`@kubernetes/sig-<group-name>-<group-suffix>`
  - <group-suffix> 可以是 bugs, feature-requests, pr-reviews, test-failures, proposals 等标签。 例如： kubernetes/sig-cluster-lifecycle-bugs, can you have a look at this?

> 这一节只提到了一些命令使用的技巧，并未对 SIG 的合作的方式做任何说明。但不可忽视的是人与人，人与组织，组织与组织的合作方式，文化氛围往往影响了整个项目的发展趋势。无论是自上而下的金字塔结构还是民主化的社区合作，构建出一套统一的运作模式才能发挥出整个团队的生产力。模式可以是随着项目的规模，阶段，市场不断变化的，但是切不可随波逐流或放任自如。

#### 分配给自己

如果你觉得你可以修复这个 issue，可以使用 `/assign` 标签将它分配给自己。如果你由于没有权限而无法分配 issue，可以在评论中留言，声明你希望认领此任务并提交 PR。

### 第五步：跟进

#### 如果一个 issue 在当前发布周期内没有任何 PR 提交

如果你看到任何已经分配到开发的 issue 在30天后还未提交任何 PR， 你应该及时联系 issue 的责任人询问他是否会提交PR或者转交给他人。

#### 如果已有 SIG 标签，但是在30天内没有任何动作

如果你发现一个 issue 已经分配了 SIG  标签，但是没有任何迹象表明它在30天内有任何行动或者讨论，请用合适的方式与 SIG 讨论该 issue 为何被搁置。 同时可以考虑参加他们的 SIG 会议，并以适当的方式提出这个问题。

#### 如果一个 issue 90 天后任然无任何活动

当它发生的时候， `fejta-bot` 机器人会为 issue 添加 `lifecycle/stale` 标签。你也可以增加 `/lifecycle frozen` 标签抢先一步来阻止机器人，或者使用 `/remove-lifecycle stale` 命令移除标签。 `fejta-bot` 会在评论中增加额外的信息。如果你没有做以上任何步骤，issue 最终会被自动关闭。

> issue 管理需要贯穿于整个开发的生命周期，形成一个闭环。及时的跟进 issue 是一个良好的工作习惯，它可帮助你在项目变得无掌控进度前进行及时的修正，保证项目有一个稳定的发布周期或按期发布。

## 备注

帮助请求 - 渠道

帮助请求类的请求需引导用户使用以下方式：

- [用户文档](https://kubernetes.io/docs/home/)和[疑难解答](https://kubernetes.io/docs/tasks/debug-application-cluster/troubleshooting/)
- [Slack](https://kubernetes.slack.com/)
- [论坛](https://discuss.kubernetes.io/)

> 以上就是原文的主要内容，工具章节并未进行翻译。有兴趣可以阅读原文。