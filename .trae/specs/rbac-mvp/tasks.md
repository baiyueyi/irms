---
topic: rbac-mvp
status: draft
---

# Tasks

## Milestone 0: Repo Skeleton

- 初始化前端（vite + vue3 + pinia + element-plus）工程结构与本地启动
- 初始化后端（golang）工程结构、配置加载与本地启动
- 连接 MySQL（使用现有 DB_* 环境变量）

## Milestone 1: Auth & Access Control

- 后端：登录接口（JWT），鉴权中间件
- 后端：must_change_password 字段与首次改密流程
- 后端：区分 SuperAdmin / User 的访问控制
- 前端：登录页、登录态存储、路由守卫
- 前端：首次登录强制改密页面与流程

## Milestone 1.5: Bootstrap

- 后端：SuperAdmin CLI 初始化（使用现有 DB_* 环境变量；检测已存在超管则退出）

## Milestone 2: Admin CRUD

- 用户管理：列表/创建/更新/启用禁用
- 用户组管理：列表/创建/更新
- 用户组成员：新增/编辑弹窗内按用户名多选并保存、查询成员
- 页面资源管理：路由扫描预览 + 同步注册 + 状态/描述维护（仅 page）
- 页面同步接口：`POST /api/pages/sync`（dry_run + apply）
- page 列表增强：返回 source 与 grant_count，keyword 支持匹配 route_path
- 页面行操作：跳转权限管理并注入 page 过滤条件
- 主机管理：列表/创建/更新/删除
- 服务管理：列表/创建/更新/删除
- 资源组管理：列表/创建/更新/删除（仅 host/service）
- 资源组成员：新增/编辑弹窗内按资源名称多选并保存（不允许 page）
- 环境管理：列表/创建/更新/删除
- 位置管理：列表/创建/更新/删除
- 主机凭据管理：列表/创建/更新/删除 + reveal
- 服务凭据管理：列表/创建/更新/删除 + reveal

## Milestone 3: Grants & Evaluation

- 授权关系：创建、列表（分页/筛选）、撤销
- 授权关系：唯一键 (subject_type, subject_id, object_type, object_id)
- 授权关系：升级/降级通过 UPDATE 完成，不产生重复有效授权
- 权限计算：对 page/host/service/credential 的有效权限聚合（ReadWrite 优先）
- service 授权前置 host ReadOnly 校验
- credential 授权前置父资源 ReadOnly 校验
- credential 动作级校验：列表仅需父资源ReadOnly；reveal需credential ReadOnly；新增/修改/删除需credential ReadWrite
- page 页面授权控制（按 page 资源对象判定）
- 权限管理 UI：主体/客体按名称搜索选择，不允许 ID 手填
- 权限管理列表：展示主体名称、客体名称、展示类型
- 权限管理 UI：新增授权显示预览文案
- 权限管理：支持编辑授权（最小能力修改 permission）
- service 环境标签继承规则（service优先，无则继承host）
- 后端：GET /api/permissions/resources 返回当前用户有效资源列表
- 前端：普通用户“我的可访问页面与资源”页面

## Milestone 4: Audit

- 关键管理操作写入审计日志
- 审计日志最小字段落库（包含 before_json/after_json、result、ip 可空）
- 审计日志查询（只读）
- reveal 操作写审计日志

## Milestone 5: E2E Validation

- 联调：完成 MVP 主链路手验
- 联调：主机/服务环境标签绑定弹窗与继承展示手验
- 构建：前端 build、后端编译检查
