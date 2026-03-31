# IRMS 后端架构级重构（Gin + Gorm + Swagger）

## 为什么必须“大改”

当前后端的主要结构性问题不是单点 bug，而是长期演进导致的“职责混杂 + 语义不统一”：

- 路由/handler/SQL/权限/审计/业务编排揉在一起（超大文件集中承载）
- DB 访问完全手写 SQL，导致接口语义与字段形状难统一，也难做事务编排与复用
- 缺少 Model / DTO / VO 明确边界，返回结构大量依赖临时 map
- Host/Service 环境绑定走碎片化接口，迫使前端维护“草稿态/已保存态”的双状态与编排 delta
- 缺少 Swagger，接口无法被规范化描述与对齐

结论：继续在旧结构上“最小修补”会持续放大前端编辑复杂度与后端维护成本，因此本轮以“结构正确 + 语义统一 + 前端可顺畅编辑”为第一目标，允许重构目录、路由层与数据访问层。

## 最终架构设计

### 分层边界

- controller：只做参数读取、绑定校验、调用 service、返回统一响应
- service：业务编排、事务、权限前置检查、审计调用、多个 repository 协作
- repository：gorm 数据访问与查询组合（禁止 controller 直接依赖 gorm DB）
- model：gorm model，对齐表结构、索引与关联关系
- dto/request：入参结构（Create/Update/Query），不在 controller 内联匿名 struct
- vo：返回给前端的结构（包含展示字段、聚合字段与来源字段）
- middleware：request_id/logger/recovery/auth/rbac/audit 等
- pkg：errors/response/paginate/validator/convert 等通用能力

### 目录结构（落地）

参考 `.trae/specs/backend-refactor-gin-gorm/spec.md`，主干结构如下：

- `backend/internal/bootstrap/*`：启动组装层
- `backend/internal/{controller,service,repository,model,dto,vo,middleware,pkg}/*`
- `backend/docs/swagger/*`：swagger 生成物（swaggo）

## gin / gorm / model / dto / vo / swagger 如何落地

### gin（HTTP）

- gin 作为唯一主 router
- 路由分组：
  - `/api/auth/*`：登录/改密等
  - `/api/*`：登录后
  - `/api/admin/*`：可选（若要将超管能力单独分组），或通过 RBAC/role middleware 控制
- middleware 统一挂载：request_id、logger、recovery、auth、rbac、audit

### gorm（数据访问）

- gorm 初始化由 bootstrap/db.go 负责，使用现有 `DB_*` 环境变量连接 MySQL
- repository 只暴露“领域相关方法”，不暴露 gorm 细节给 controller
- service 中统一做事务（`db.Transaction(func(tx *gorm.DB){...})`）
- 列表查询避免 N+1：优先 join/preload/批量查询

### model（数据模型）

- 建立与现有表结构对齐的 gorm model（例如 Host/Service/Environment/Location/Page/Grant/...）
- 明确表名、字段映射、索引、关联关系
- 重要关系表：
  - host_environments
  - service_environments

### dto/request（入参）

- HostUpdateRequest / ServiceUpdateRequest 为 typed request
- Create/Update/Query 分开定义
- controller 统一使用 gin binding/validator 完成校验

### vo（返回结构）

- HostVO：包含环境 ids/codes/names 与 location_name 等展示字段
- ServiceVO：区分自身/继承/生效环境，并给出 environment_source
- VO 由 service 组装，避免 controller 拼装

### Swagger（swaggo）

- 使用 swaggo 方案：
  - 通过注释标注 controller 的路由、参数、返回
  - 生成 swagger 文档到 `backend/docs/swagger`
  - gin 中挂载 swagger UI 路由

## Host / Service 编辑接口如何重构（editDraft 核心）

### 设计原则

- 前端编辑只维护一个 `editDraft`（完整对象深拷贝）
- 保存时尽量一次提交“统一对象”，后端负责落库与关系表编排
- 环境标签更新不再要求前端逐条 POST/DELETE 编排

### 接口（建议）

- `PUT /api/hosts/:id`
  - 入参：HostUpdateRequest（typed，含 `environment_ids`）
  - 行为：事务内更新 hosts + replace host_environments
  - 返回：HostVO（更新后的最新数据）

- `PUT /api/services/:id`
  - 入参：ServiceUpdateRequest（typed，含 `environment_ids`）
  - 行为：事务内更新 services + replace service_environments
  - 返回：ServiceVO（更新后的最新数据）

### environment_ids 如何纳入统一对象提交

- request 直接携带 `environment_ids: []`
- service 层依据 `environment_ids` 执行 replace：
  - 删除该 host/service 现有关系
  - 批量插入新关系（去重后）
  - 事务提交保证一致性

## page route_path 与 /admin 前缀去除如何处理

### canonical 规则

- 只允许以 `/` 开头的绝对路由
- 统一以“去掉 /admin 前缀后的路径”作为 canonical route_path
- 权限判断与页面资源同步都基于 canonical route_path

### 迁移方案

- 提供 CLI 迁移命令（幂等）：
  - 将 `pages.route_path` 中 `/admin/...` 更新为 `/...`
  - 如遇唯一键冲突，跳过并输出冲突报告
- 必要时同步修复 `resources.route_path`（若存在 page 类型资源）

## 兼容策略（避免前端与已有数据瞬间失效）

### 路由迁移（chi -> gin）

- gin 作为主入口
- 旧 chi router 以 `legacy` 方式挂载（fallback）：
  - 新实现的路由由 gin 直接处理
  - 未迁移的旧路由由 legacy handler 继续处理
- 迁移阶段可通过日志/指标统计哪些旧接口仍被调用，逐步替换

### SQL -> gorm

- 新接口全部使用 gorm repository
- 旧接口保留直到被 gin 新 controller 替换

### 接口兼容

- 优先保持原路径不变（避免前端改动）
- 若必须变更语义：
  - 先保留旧接口在 legacy 层
  - 新接口提供等价能力并完善 swagger
  - 前端切换后再下线旧接口

