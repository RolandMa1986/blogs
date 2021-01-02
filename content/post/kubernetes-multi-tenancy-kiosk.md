---
title: "Kiosk 多租户扩展-基本概念"
date: 2020-12-30T20:46:17+08:00
---

## Kubernetes 多租户扩展

- 使用 Accounts 与 Account User 在共享的 Kubernetes 集群中区分租户
- 租户自助 Namespace 空间配置
- Account Limits 账户配额保障服务质量与公平性
- Namespace Templates 命名空间模板保证租户隔离与自助初始化 Namespace。
- *多集群租户管理以共享集群池

## Why kiosk

Kubernetes 被设计为一个单租户平台，因此在单集群上实现多租户功能变得十分困难。然而，共享一个集群有很多优势，例如，资源利用率高，管理配置成本低，也可使租户更容易共享集群内部资源。

当然，创建多租户 Kubernetes 集群的方法有很多，许多 Kubernetes 发行版也提出了他们自己的租户逻辑，然而却没有一个种轻量级的，可热拔插，并可定制化的解决方案。使管理员可以在任何标准的 Kubernetes 集群之上增加多租户能力。

## kiosk 愿景

- 100% 开源：使用 CNCF 兼容的 Apache 2.0 license
- 即插即用：可轻而易举地在任何已有集群中安装，适用于多种不同的场景
- 快速：着重于租户的自动化与自助服务
- 安全：为不同级别的租户隔离提供默认配置
- 可扩展：providing building blocks for higher-level Kubernetes platforms

## 架构

kiosk 的核心理念是使用 Kubernetes Namespace 作为隔离的工作空间，租户的应用可以运行在彼此隔离的空间中。为了降低管理成本，集群管理员应该配置 kiosk，然后它将成为一个租户可以在 Kubernetes 中自助配置 Namespace 的系统。

## 工作流程与交互

下图展示了kiosk中的主要参与者和 Kubernetes 资源以及他们之间的关系。

![kiosk Workflow](https://github.com/kiosk-sh/kiosk/blob/master/docs/website/static/img/kiosk-workflow-kubernetes-multi-tenancy-extension.png)

#### Cluster Admin

集群管理员具有集群集群级别、非命名空间资源的CRUD操作权限（类似于 RBAC 相关资源一样的自定义资源，如，Account, AccountQuota, AccountQuotaSet 和 Template）。集群管理员通过创建或管理 Accounts，AccountQuotas， AccountQuotaSet 和 Template 配置 kiosk。

#### Account

每个租户使用一个 Account 表示。集群管理员定义和管理账户 Account，并将用户（User，Groups，ServiceAccounts）分派到账户中，这个方式类似于 RBAC 中使用 Rolebinding 将绑定 Role 与 subject 关联类似。

#### Account User

Account User 使用特定的账户调用 API Server 执行 Kubernetes 操作。管理员可以将一个 Account User 分派到多个 Account （租户）中。 Account User 可以访问 Account 拥有的 Space 空间。如果分配了默认的 kiosk ClusterRole，每个Account User 都将拥有 Account 的 list、get、create、delete 等 Space 权限， 当然这些权限可以通过 RBAC RoleBinding 修改。

#### Space

Space 是非持久化的、虚拟的资源，用于表示一个 Kubernetes Namespace。Space 有如下特征：
- 每个 Space 可以归属于一个 Account，即拥有者，Space 也可能没有拥有者。
- 如果一个用户拥有访问对应 Namespace 的权限， 那么此用户同样可以访问 Space。因此，除了 Account User 外，其他参与者（User、Group、ServiceAccount）假如已经通过 Kubernetes RBAC 进行了授权，那么它们也可以访问 Space。
- 每个用户能且仅能看到他们拥有权限的 Space。与之形成对照的是常规的 Namespace， 用户只能被授予查看所有 Namespace 或者一个也无法看见。
- Space 的拥有权可以更改，仅需要修改 Namespace 上的 annotation。
- Space 在创建过程中(或者修改拥有权时)会在 Space 对应的 Namespace 下为 Account 创建一个 RoleBinding。相应的 RBAC ClusterRole 可以在 Account 中配置。
- Space 在创建时，可以通过在 Account 下配置默认 Template 模板，对其进行一系列的资源填充。 Kiosk 保证这些资源会在用户获取 Namespace 访问权之前被创建。

#### Namespace

Namespace 是一个常规的 Kubernetes 空间，任何拥有对应 RBAC 权限的用户都可对其进行访问。Kiosk会对 Namespace 进行配置与管理，并保证其与 Space 是一对一的关系。默认情况下，Account User 拥有对其账户 Account 下的所有的 Space 对应的 Namespace 的操作权。

#### Template

集群管理员可以定义、管理 Template。 Template 用于初始化 Space/Namespace 中的一系列 Kubernetes 资源(使用 manifest 或者 Helm Chart)。Template 能使用不同于 Account User 权限的ClusterRole 创建，因此他们可以创建 Space/Namespace 的参与者所不允许的资源，例如，设置资源隔离（Network Polices，Pod Serurity Polices)。集群管理员可以在 Account 中创建多个默认 Template,他们会被自动的应用在 Account 下的每一个 Space。除此，Account User 可以创建非强制性的 Template， 应用于新建的 Space。

#### TemplateInstance

将 Template 应用于 Space 时， Kiosk 会创建一个 TemplateInstance 用于追踪哪些 Template 已经作用于 Space。TemplateInstance 即保存了 Template 信息，也保存了初始化的相关参数。除此以外，TemplateInstance 可以配置保持 Template 同步，例如，当 Template 发生变化时，TemplateInstance 也会更新相关资源。

#### AccountQuota

集群管理员可以创建并管理 AccountQuota。AccountQuota 在集群级别定义了 Account 的总限额。所有归属于 Account 的 Space/Namespacce 的资源总和限制被定义在 AccountQuota 中。与 Namespace 可以应用多个 ResourceQuotas 类似，一个 Account 也可以受多个 AccountQuotas 限制。如果多个 AccountQuotas 中包含有同种限额（如，CPU），那么将使用最低的限额作用于 Account。

## 定制资源和资源组

在 Kubernetes 集群中安装 Kiosk 后，以下组件将被添加到集群中：

- Account, AccountQuota, AccountQuotaSet, Template, TemplateInstance 等资源的CRD
- kiosk 资源的控制器（集群内运行）
- API Server 扩展（同控制器一样在集群内运行）

![kiosk Resources](https://github.com/kiosk-sh/kiosk/blob/master/docs/website/static/img/kiosk-data-structure-kubernetes-multi-tenancy-extension.png)

Kiosk 在 Kubernetes 标准 API Groups 中增加了两种资源组：
1. **Custom Resources: config.kiosk.sh**
  用于配置 kisok 的 自定义资源(CRDs)。这些资源与其他资源一样被持久化在 etcd 中保存，通过一个运行在集群中的 operator 进行管理。

      - config.kiosk.sh/Account
      - config.kiosk.sh/AccountQuota
      - config.kiosk.sh/AccountQuotaSet (soon)
      - config.kiosk.sh/Template
      - config.kiosk.sh/TemplateInstance

2. **API Extension: tenancy.kiosk.sh**
  虚拟资源可以通过 API Server 扩展进行访问，但它们并不会保存在 etcd 中。这些资源类似于传统数据库中的视图。提供这些虚拟资源而不仅仅使用 CRDs 的好处是我们可以为每个请求动态的计算访问权限。同时也意味它不仅能提供获取，编辑和管理 Space 的能力，它也允许我们能够根据 Account User 所属的 Account 来展示其所拥有的 Space。换句话它打破了 Kubernetes 仅能在集群级别授予访问权限，再进行过滤的限制。

      - tenancy.kiosk.sh/Account
      - tenancy.kiosk.sh/AccountQuota
      - tenancy.kiosk.sh/Space
      - tenancy.kiosk.sh/TemplateInstance

