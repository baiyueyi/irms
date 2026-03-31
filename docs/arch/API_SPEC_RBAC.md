# IRMS RBAC API Spec (MVP)

## 1) 约定

- API 前缀统一：`/api/...`
- 不使用：`/api/v1`、`/api/v2`
- 鉴权：Bearer JWT（`Authorization: Bearer <token>`）
- CLI 初始化超管为命令行能力，不是 HTTP 接口
- 登录账号：`username + password`
- 首次登录强制改密：`must_change_password=true` 时，仅允许访问改密与基础会话接口
- 业务边界：`page` 仅表示前端页面/路由访问资源，不与 host/service 混合管理

## 2) 通用响应结构

成功：

```json
{
  "code": "OK",
  "message": "success",
  "data": {}
}
```

失败：

```json
{
  "code": "INVALID_ARGUMENT",
  "message": "参数错误",
  "details": {}
}
```

## 3) 错误码

- `OK`
- `UNAUTHORIZED`
- `FORBIDDEN`
- `FIRST_LOGIN_PASSWORD_CHANGE_REQUIRED`
- `INVALID_ARGUMENT`
- `NOT_FOUND`
- `CONFLICT`
- `RESOURCE_TYPE_MISMATCH`
- `PRECONDITION_REQUIRED`（前置权限不足）
- `INTERNAL_ERROR`

## 4) 分页与筛选约定

- Query:
  - `page`（默认 1）
  - `page_size`（默认 20，最大 100）
  - `keyword`（名称关键字）
- 分页响应：

```json
{
  "list": [],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 0
  }
}
```

## 5) Auth

- `POST /api/auth/login`
- `POST /api/auth/change-password`
- `GET /api/me`

## 6) Users / Groups（SuperAdmin）

- `GET/POST/PUT/DELETE /api/users`
- `GET/POST/PUT/DELETE /api/user-groups`
- `GET/POST/DELETE /api/user-group-members`

管理端交互要求：
- 新增/编辑用户组时可按用户名多选成员
- 用户组列表显示成员数，并支持查看成员名称

## 7) Pages（页面资源，SuperAdmin）

- `GET /api/pages`：列表筛选 `keyword,status`（keyword 匹配 `name` 或 `route_path`）
- `POST /api/pages/sync`：路由同步（支持 `dry_run`）
  - 请求：`{ dry_run: boolean, routes: [{ name, route_path, source }] }`
  - 响应：`summary,new_routes,existing_routes,changed_routes,retired_routes`
- `POST /api/pages`：补录异常路由（source 默认 `manual`）
- `PUT /api/pages/{id}`：`name,route_path,source,status,description`
- `DELETE /api/pages/{id}`

说明：
- `page` 只表示页面访问资源，例如 `/admin/users`、`/admin/grants`、`/my-resources`
- page 主来源是前端 router/menu 同步注册，不是默认手工 CRUD 对象
- `/admin/pages` 为页面资源管理正式入口；`/admin/resources` 仅做兼容跳转
- page 不承载 host/service 业务属性，不绑定 location

## 8) Hosts（SuperAdmin）

- `GET /api/hosts`：筛选 `keyword,status,provider_kind,location_id,environment_id`
- `POST /api/hosts`
- `PUT /api/hosts/{id}`
- `DELETE /api/hosts/{id}`
- `GET /api/hosts/{id}/services`：主机下服务列表
- `GET/POST/DELETE /api/host-environments`：host 与 environment 多对多绑定

字段：
- `id,name,hostname,primary_address,provider_kind,cloud_vendor,cloud_instance_id,os_type,status,location_id,description`

## 9) Services（SuperAdmin）

- `GET /api/services`：筛选 `keyword,status,service_kind,host_id,environment_id`
- `POST /api/services`
- `PUT /api/services/{id}`
- `DELETE /api/services/{id}`
- `GET/POST/DELETE /api/service-environments`：service 与 environment 多对多绑定

字段：
- `id,name,service_kind,host_id,endpoint_or_identifier,port,protocol,cloud_vendor,cloud_product_code,status,description`

规则：
- `service.host_id` 可为空（云服务可不绑定主机）
- service 环境标签生效：
  - service 有标签：使用 service 标签
  - service 无标签且 host_id 不为空：继承 host 标签
  - service 无标签且 host_id 为空：无环境标签

## 10) Environments（SuperAdmin）

- `GET/POST/PUT/DELETE /api/environments`
- 字段：`id,code,name,status,description`

## 11) Locations（SuperAdmin）

- `GET/POST/PUT/DELETE /api/locations`
- 字段：`id,code,name,location_type,address,status,description`

规则：
- host 绑定 location（一对一）
- service/page 不允许直接绑定 location

## 12) Resource Groups（SuperAdmin，Compatibility Storage）

- `GET/POST/PUT/DELETE /api/resource-groups`
- `GET/POST/DELETE /api/resource-group-members`

规则：
- 资源组仅允许 `host`、`service` 两种类型
- `page` 不允许进入资源组模型
- 管理端资源组 type 下拉仅展示 `host`、`service`
- 管理端显示语义类型：`host_group`、`service_group`
- 新增/编辑资源组时可按名称多选资源成员
- `host_group` 仅可选 host；`service_group` 仅可选 service
- `resource_groups/resource_group_members` 为 compatibility storage；对外语义统一为 `host_group/service_group`

## 13) Credentials（Authenticated + RBAC）

### 13.1 Host Credentials

- `GET /api/host-credentials?host_id=...`（列表不返回明文）
- `POST /api/host-credentials`
- `PUT /api/host-credentials/{id}`
- `DELETE /api/host-credentials/{id}`
- `POST /api/host-credentials/{id}/reveal`（查看明文，写审计）

### 13.2 Service Credentials

- `GET /api/service-credentials?service_id=...`（列表不返回明文）
- `POST /api/service-credentials`
- `PUT /api/service-credentials/{id}`
- `DELETE /api/service-credentials/{id}`
- `POST /api/service-credentials/{id}/reveal`（查看明文，写审计）

规则：
- 密文使用 `CREDENTIAL_ENCRYPTION_KEY` 加密存储
- `certificate_pem_ciphertext/private_key_pem_ciphertext` 使用 `LONGTEXT` 存储密文 PEM
- 列表接口不得返回明文 `secret/certificate/private_key/passphrase`
- 动作级校验最小规则：
  - 前置：必须先有父资源 `host.read`（host）
  - 列表（GET）：仅要求父资源 `host.read`
  - 明文查看（POST /reveal）：要求 `host_credential.read`（`host_credential.write` 隐含）
  - 新增/修改/删除：要求 `host_credential.write`
- service 凭据动作级最小规则：
  - 前置：必须先有父资源 `service.read`（service）
  - 列表（GET）：仅要求父资源 `service.read`
  - 明文查看（POST /reveal）：要求 `service_credential.read`（`service_credential.write` 隐含）
  - 新增/修改/删除：要求 `service_credential.write`
- 凭据权限码仅保留：`host_credential.read`、`host_credential.write`、`service_credential.read`、`service_credential.write`
- 不再存在权限码：`host_credential.reveal`、`service_credential.reveal`

## 14) Grants（SuperAdmin）

- `GET /api/grants`
- `POST /api/grants`
- `PUT /api/grants/{id}`（最小编辑能力：更新 permission）
- `DELETE /api/grants/{id}`

授权对象类型：
- `page`
- `host`
- `service`
- `host_credential`
- `service_credential`
- `host_group`
- `service_group`

说明：
- 管理端授权 UI 不使用手填 `subject_id/object_id`
- 授权创建流程必须先按名称搜索并选择主体/客体，再由前端提交对应 ID
- grants 列表必须返回并展示主体名称、客体名称，不以裸 ID 作为主要交互
- 管理端授权 UI 支持新增/编辑/删除
- 新增授权弹窗需展示“主体-客体-权限”预览文案

授权规则：
- service 授权前置：当 `service.host_id` 非空，授予 `service.read/service.write` 前必须已有对应 host 的有效 `host.read` 以上权限
- credential 权限前置：授予/生效 host_credential 或 service_credential 权限前，必须先有对应父资源 `host.read/service.read` 以上权限
- 权限管理页（/admin/grants）对凭据对象仅展示 `read/write` 两档，不展示 reveal 权限码

## 15) Permissions

- `GET /api/permissions/resources`：返回当前用户可访问页面资源（page）与权限
- `GET /api/permissions/services`：返回当前用户可访问服务资源与权限（含环境生效结果）
- `GET /api/permissions/hosts`：返回当前用户可访问主机资源与权限

## 16) Audit Logs（SuperAdmin）

- `GET /api/audit-logs`

审计最小要求：
- 凭据明文查看（reveal）
- 凭据新增/修改/删除
- 授权新增/撤销
- 授权更新（permission 变更）

审计动作与目标说明（最小基线）：
- 凭据：`create_host_credential`、`update_host_credential`、`delete_host_credential`、`reveal_host_credential`、`create_service_credential`、`update_service_credential`、`delete_service_credential`、`reveal_service_credential`
- 授权：`upsert_grant`、`update_grant`、`revoke_grant`
- 目标类型使用正式语义：`host_group`、`service_group`；不再以 `resource_group` 作为正式对外语义

## 16.1 凭据动作行为矩阵

| 凭据动作 | 接口 | 父资源前置 | 凭据权限码 |
|---|---|---|---|
| 列表主机凭据 | `GET /api/host-credentials?host_id=...` | `host.read` | 无 |
| 新增主机凭据 | `POST /api/host-credentials` | `host.read` | `host_credential.write` |
| 修改主机凭据 | `PUT /api/host-credentials/{id}` | `host.read` | `host_credential.write` |
| 删除主机凭据 | `DELETE /api/host-credentials/{id}` | `host.read` | `host_credential.write` |
| 查看主机凭据明文 | `POST /api/host-credentials/{id}/reveal` | `host.read` | `host_credential.read`（或 `host_credential.write`） |
| 列表服务凭据 | `GET /api/service-credentials?service_id=...` | `service.read` | 无 |
| 新增服务凭据 | `POST /api/service-credentials` | `service.read` | `service_credential.write` |
| 修改服务凭据 | `PUT /api/service-credentials/{id}` | `service.read` | `service_credential.write` |
| 删除服务凭据 | `DELETE /api/service-credentials/{id}` | `service.read` | `service_credential.write` |
| 查看服务凭据明文 | `POST /api/service-credentials/{id}/reveal` | `service.read` | `service_credential.read`（或 `service_credential.write`） |

## 17) CLI 初始化（非 HTTP）

- 命令职责：创建首个 SuperAdmin
- 输入：命令行交互输入 `username`、`password`
- DB：使用现有 `DB_*` 环境变量连接 MySQL
- 规则：
  - 已存在 SuperAdmin 时直接退出
  - 不自动创建默认账号
  - 不提供首次注册 HTTP 接口
