# Tasks

## T1 审计现状

- 梳理现有 chi 路由/handler 结构、SQL 分布、schema/migrate、Auth/RBAC/审计现状
- 列出现有 Host/Service/Environment/Page 关键接口与数据形状

## T2 设计与兼容策略

- 明确新分层目录与依赖边界
- 定义统一响应与错误码策略
- 定义 Host/Service 更新语义与 VO 形状
- 定义 legacy 兼容方式（gin 主路由 + fallback 旧 handler）
- 定义 page route_path canonical 规则与迁移策略

## T3 基础设施落地

- 引入 gin、gorm、swagger 依赖
- 新增 bootstrap（app/db/router/swagger）
- 新增统一 middleware（request_id/logger/recovery/auth）
- 新增 pkg/response 与 pkg/errors

## T4 领域分层落地（核心资源优先）

- model：对齐现有 schema（hosts/services/environments/locations/pages/grants/users/user_groups/credentials/…）
- repository：gorm CRUD + 典型查询（分页、过滤、批量 preload）
- service：事务、权限检查、审计编排
- controller：绑定/校验/返回统一响应

## T5 Host/Service 编辑接口重构

- `PUT /api/hosts/:id` typed request + replace host_environments（事务）
- `PUT /api/services/:id` typed request + replace service_environments（事务）
- 可选：保留 `PUT /api/*/:id/environments` replace 风格接口

## T6 VO 列表与环境来源规则

- `GET /api/hosts` 返回 HostVO（环境 codes/names）
- `GET /api/services` 返回 ServiceVO（own/inherited/effective/source）

## T7 Swagger

- swaggo 生成 docs
- gin 中挂载 swagger 路由并验证可访问

## T8 Page route_path 规范化与迁移

- 迁移命令/脚本：`/admin/...` -> `/...`（冲突可报告/跳过）
- 同步与权限判断统一使用 canonical route_path

## T9 验证与报告

- `go build ./...`
- 运行服务并验证：auth、hosts、services、environments、pages、grants、credentials（最小闭环）
- 输出重构报告与迁移执行说明（含 swagger 生成与访问地址）

