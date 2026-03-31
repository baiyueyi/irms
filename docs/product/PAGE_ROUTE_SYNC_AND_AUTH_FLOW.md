# 页面资源同步与权限命中说明

## 1. 范围

- page 只表示前端页面/路由资源。
- page 不属于 host/service 业务资源域，不进入资源组。
- 页面授权唯一入口是“权限管理”。

## 2. 页面资源来源

- 主来源：前端 router/menu。
- 页面资源管理页（`/admin/pages`）负责：
  - 扫描路由候选项；
  - 预览差异（新发现、已存在、信息变化、已下线）；
  - 管理员确认后执行同步。
- 手工创建仅保留“补录异常路由”场景，source=manual。

## 3. 页面资源管理职责

- 同步注册 page（`POST /api/pages/sync`，支持 dry_run）。
- 维护 page 状态与描述（active/inactive）。
- 提供“去授权”跳转到权限管理并带 page 过滤参数。
- 不在该页直接创建 grant。

## 4. 路由与入口边界

- 正式入口：`/admin/pages`。
- 兼容入口：`/admin/resources` 仅重定向到 `/admin/pages`。
- 左侧菜单文案保持“页面资源管理”。

## 5. 权限命中规则

- route_path 是 page 识别与命中核心字段之一。
- 普通用户页面访问由 `GET /api/permissions/resources` 返回的 page `route_path` 判定。
- 仅当 `route_path` 命中可访问集合时允许进入对应页面。

## 6. credential 权限口径（本轮收口）

- 权限管理页对凭据对象仅展示两档权限：`read`、`write`。
- 对应权限码：
  - `host_credential.read` / `host_credential.write`
  - `service_credential.read` / `service_credential.write`
- 不再展示或写入 `host_credential.reveal`、`service_credential.reveal` 权限码。
- 所有 credential 动作仍受父资源 `host.read/service.read` 前置约束。
- 明文查看必须走 reveal 接口并写审计日志；reveal 由 `*.read` 权限覆盖。

## 7. resources 定位（兼容层）

- `resources/resource_groups/resource_group_members` 仅保留为 compatibility storage。
- 正式对外语义与权限管理展示使用：`page`、`host`、`service`、`host_group`、`service_group`。
- `/admin/resources` 仅作为历史入口兼容跳转，不承载新的领域能力。
