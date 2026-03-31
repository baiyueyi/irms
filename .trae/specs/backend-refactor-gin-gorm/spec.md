# Backend Refactor：Gin + Gorm + Swagger（架构级重构）

## 目标

- 后端实际使用 gin 作为 HTTP 框架，路由与中间件结构清晰
- 后端实际使用 gorm 进行数据访问，禁止在 controller/handler 里手写 SQL
- 明确分层：controller / service / repository / model / dto(request) / vo
- API 语义统一：路径、方法、响应形状、错误码、参数校验方式一致
- Swagger 可生成、可访问，覆盖主要资源域
- 支撑前端“单一 editDraft”编辑模型：
  - 列表接口返回 VO（包含环境信息与来源规则）
  - Host/Service 更新接口使用 typed request，允许一次性提交 environment_ids 并在后端事务内 replace 关系表

## 非目标（本轮不做或仅做兼容）

- 不引入复杂策略引擎、多租户、审批流
- 不做字段级权限与审计平台化抽象
- 不强制改变现有数据库 schema；如需调整，必须提供迁移方案与可回滚说明

## 架构设计（落地形态）

### 目录结构（主干）

- backend/cmd/irms/
  - main.go（启动 HTTP + CLI 子命令）
- backend/internal/bootstrap/
  - app.go（组装：config/db/router/swagger）
  - router.go（gin 路由与分组）
  - db.go（gorm 初始化）
  - swagger.go（swagger 路由挂载）
- backend/internal/model/（gorm models，对齐现有表结构）
- backend/internal/dto/request/（Create/Update/Query 请求入参）
- backend/internal/vo/（返回给前端的 VO）
- backend/internal/repository/（gorm 数据访问）
- backend/internal/service/（业务编排 + 事务 + 权限检查 + 审计）
- backend/internal/controller/（参数绑定/校验 + 调用 service + 统一响应）
- backend/internal/middleware/（request_id/logger/recovery/auth/rbac/audit）
- backend/internal/pkg/（errors/response/paginate/validator 等通用能力）

### API 统一规范

- API 前缀统一 `/api/...`
- 统一响应结构：
  - `code`：稳定错误码/状态码（字符串）
  - `message`：面向前端的短消息（不直接透出内部 err）
  - `data`：业务数据（对象/列表/分页）
  - `request_id`：链路追踪字段
  - `details`：可选（仅用于参数校验等可公开信息）

### Host / Service 更新语义（editDraft 支撑）

- `PUT /api/hosts/:id`：接收 `HostUpdateRequest`（typed），其中 `environment_ids` 可选
  - service 层事务内：
    - 更新 hosts 主表
    - replace `host_environments`
- `PUT /api/services/:id`：接收 `ServiceUpdateRequest`（typed），其中 `environment_ids` 可选
  - service 层事务内：
    - 更新 services 主表
    - replace `service_environments`

### Service 环境规则（后端输出，前端不拼）

- `own_environment_*`：service 自身绑定
- `inherited_environment_*`：当 `service` 没有自身环境且 `host_id != null` 时，继承 host 环境
- `effective_environment_*`：最终生效环境
- `environment_source`：`own|host_inherited|none`

### Page route_path 规范化

- page 的 `route_path` 视为 canonical path（权限判断关键字段）
- 兼容旧数据：
  - 提供 CLI 迁移命令，将 `/admin/...` 规范化到 `/...`
  - 同步接口/鉴权逻辑基于 canonical path

## 迁移与兼容策略

- 服务从 chi 迁移到 gin：
  - 以 gin 作为主 router
  - 对于尚未迁移的旧接口，提供 `legacy` 兼容层（在 gin 路由末尾 fallback 到旧的 http.Handler），确保前端与既有能力不瞬间失效
- 数据访问从手写 SQL 迁移到 gorm：
  - 新接口全部走 gorm repository
  - 旧接口保留原实现直至被替换
- 接口兼容：
  - 新接口优先保持路径不变（/api/hosts、/api/services 等），以“新实现覆盖旧实现”的方式逐步迁移
  - 旧环境绑定碎片接口允许保留一段时间，但不再是前端主编辑流程的唯一方式

## 验收标准（优先）

- 后端实际使用 gin + gorm，且分层目录存在并可读
- Host/Service 编辑接口可接收统一对象，并支持 `environment_ids` 一次性 replace
- 列表接口返回 VO，不返回 map 作为主要结构
- Swagger 可生成并可打开
- `go build ./...` 通过，服务可运行并能验证核心接口

