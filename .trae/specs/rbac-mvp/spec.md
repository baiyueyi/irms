---
topic: rbac-mvp
status: draft
---

# IRMS RBAC MVP Spec

## Goal

打通“用户、用户组、页面资源、主机、服务、环境、位置、凭据、授权关系”主链路，交付可用的 Web 管理后台与后端 API，并确保授权判定生效与可撤销、可查询。

## Success Criteria (MVP)

- 超管可在后台完成：用户/用户组/page/host/service/environment/location/credential/授权的增删改查（最小必要字段）。
- 普通用户无法访问后台管理页面与对应 API。
- 授权判定对 page/host/service/credential 生效，撤销后立即失效；同时命中时 ReadWrite > ReadOnly。
- 前端可构建：vite + vue3 + pinia + element-plus。
- 后端可运行：golang + mysql；Windows 本地开发与联调可用。

## Non-Goals

- 不做审批/流程/策略引擎/字段级权限/多租户/移动端/报表/消息中心/外部系统集成。

## Roles

- SuperAdmin：管理后台全量权限。
- User：仅可访问“自己有权限的 page 及业务资源范围”（MVP 以接口/页面占位方式体现）。

## RBAC Model

Entities:

- User
- UserGroup
- UserGroupMember (User ↔ UserGroup)
- Page
- Host
- Service
- ResourceGroup
- ResourceGroupMember
- Environment
- Location
- HostCredential
- ServiceCredential
- HostEnvironment (Host ↔ Environment)
- ServiceEnvironment (Service ↔ Environment)
- Grant
- AuditLog

PermissionSet:

- ReadOnly
- ReadWrite

Grant dimensions:

- Subject: User | UserGroup
- Object: page | host | service | host_group | service_group
- Permission: ReadOnly | ReadWrite

Decision rules:

- 计算来源：用户直授、用户组直授（用户/用户组 × page/host/service/credential）。
- 命中任一有效授权视为有权限；多命中时取更高权限（ReadWrite）。
- service 授权前置：当 service.host_id 不为空时，授予 service 权限前必须先有对应 host 的有效 ReadOnly。
- credential 权限前置：授予 host_credential/service_credential 前，必须先有对应父资源的有效 ReadOnly。
- credential 动作级策略：列表=父资源ReadOnly；明文查看=父资源ReadOnly+credential ReadOnly；新增/修改/删除=父资源ReadOnly+credential ReadWrite。

## Domain Constraints

- page 仅表示前端页面/路由资源，不参与 host/service 业务建模。
- page 主来源为前端 router/menu 同步；手工新增降级为异常补录能力。
- 页面资源管理负责路由扫描差异（新增/变更/下线）并确认同步，授权仍在权限管理完成。
- route_path 是 page 资源唯一标识与页面权限命中的核心字段。
- host 与 service 为独立业务资源，不与 page 混在同一管理页。
- 资源组仅允许 host_group/service_group（底层 type: host|service），不允许 page 进入资源组。
- service.host_id 可为空；云产品允许无主机关联。
- host 与 location：一个 host 只能绑定一个 location。
- host 与 environment：多对多；service 与 environment：多对多。
- service 环境继承规则：
  - service 有标签：使用 service 自己标签
  - service 无标签且 host_id 非空：继承 host 标签
  - service 无标签且 host_id 为空：无标签
- credential 独立表存储，不放在 host/service 主表。
- 证书/私钥密文字段使用 LONGTEXT。

## API Conventions

- 统一前缀：/api/...
- 不使用 /api/v1 或 /api/v2。

## Environment Variables

- 只使用现有环境变量（不引入 dotenv，不覆盖已有变量名/值）：
  - CREDENTIAL_ENCRYPTION_KEY
  - JWT_SECRET
  - DB_HOST
  - DB_NAME
  - DB_PASSWORD
  - DB_USER
  - DB_PORT

## MVP Pages (Admin)

- 登录页
- 用户管理
- 用户组管理
- 用户组新增/编辑包含成员多选（按用户名）
- 用户组列表显示成员数并支持查看成员名称
- 页面资源管理（仅 page，同步注册模式）
- 主机管理
- 服务管理
- 资源组管理（仅 host/service）
- 资源组新增/编辑包含成员多选（按资源名称，按类型联动）
- 环境管理
- 位置管理
- 权限（授权）管理
- 权限管理新增授权包含“主体-客体-权限”预览
- 权限管理支持编辑授权（最小能力：修改权限）

## MVP Pages (User)

- 我的可访问页面与资源

## MVP API Surface (Draft)

Auth:

- POST /api/auth/login
- POST /api/auth/logout (optional)
- GET /api/me

Password:

- POST /api/auth/change-password

Admin CRUD:

- /api/users
- /api/user-groups
- /api/user-group-members
- /api/pages
- /api/hosts
- /api/services
- /api/resource-groups
- /api/resource-group-members
- /api/environments
- /api/locations
- /api/host-environments
- /api/service-environments
- /api/host-credentials
- /api/service-credentials
- /api/grants
- /api/audit-logs (read-only)

Permission evaluation:

- GET /api/permissions/resources (当前用户可访问 page 资源列表/范围)

## Data (Draft Constraints)

- users.username 唯一
- users.must_change_password 用于首次登录强制改密
- user_groups.name 唯一（可选，若名称作为业务键）
- pages.route_path 唯一
- hosts.location_id 可空且单值
- services.host_id 可空
- host_environments (host_id, environment_id) 唯一
- service_environments (service_id, environment_id) 唯一
- host_credentials / service_credentials 使用密文字段保存敏感信息
- user_group_members (user_id, user_group_id) 唯一
- grants 唯一键：(subject_type, subject_id, object_type, object_id)
- 权限管理交互不允许手填 subject_id/object_id，必须按名称搜索并选择
- grants.permission 不在唯一键内，取值固定：ReadOnly | ReadWrite
- grants 权限升级/降级通过 UPDATE 完成
- grants 撤销授权：删除当前有效 grant 记录
- grants 不保留重复历史，历史通过 audit_logs 保留

## Audit (MVP)

至少记录关键管理操作：

- 用户/用户组/page/host/service/environment/location：创建、更新、禁用/启用（如支持）
- 用户组成员、环境标签绑定：添加、移除
- 资源组/资源组成员：创建、更新、添加、移除（仅 host/service）
- credential 接口对已登录用户开放，由 RBAC 动作级校验决定是否允许执行
- 授权：创建、撤销
- 凭据 reveal（明文查看）：必须记录审计

## Auth Decisions

- 登录形态：username + password
- 首次登录强制改密：users.must_change_password = true 时必须改密后才允许继续使用系统
- SuperAdmin 初始引导：通过后端 CLI 初始化命令创建；检测到已存在超管后必须退出且不重复创建

## Audit Fields (Minimum)

- id
- actor_user_id
- actor_username_snapshot
- action
- target_type
- target_id
- target_name_snapshot
- occurred_at
- before_json
- after_json
- result
- ip（可空）
