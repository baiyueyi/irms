# Release Note：Backend Refactor（Gin + Gorm + Swagger）

## 目标

- 后端结构正确：gin 路由 + gorm 数据访问 + 明确分层
- API 语义更统一：typed request + 统一响应形状
- 支撑前端单一 editDraft：Host/Service 更新可一次提交 environment_ids 并由后端事务内 replace
- Swagger 可用：可生成、可访问、可联调

## 关键变更

- 新增 `internal/bootstrap`：统一组装启动链路（gin engine / swagger / middleware）
- 新增 `internal/{model,dto,vo,repository,service,controller,middleware,pkg}`：落地分层边界
- Gin 主路由作为唯一 `/api` 入口：
  - 已移除 `NoRoute -> legacy.ServeHTTP` 兜底，避免长期双轨黑箱
- Host/Service 编辑接口统一：
  - `PUT /api/hosts/:id`、`PUT /api/services/:id` 接收 typed request
  - 支持 `environment_ids` 一次提交并 replace 关系表
  - `environment_ids: []` 表示显式清空环境绑定（合法）
  - `POST /api/hosts`、`POST /api/services` 也支持 `environment_ids` 单对象提交
  - HostVO/ServiceVO 明确输出 own/inherited/effective 环境与 source/label
- 安全动作审计：
  - `POST /api/auth/change-password` 成功/失败均写审计（不记录明文密码）
- JSON 契约更严格：
  - 关键写接口启用严格 JSON 解码（未知字段返回 400，避免静默吞字段）
- 读接口契约收紧：
  - 关键列表接口的 query 参数（如 page/page_size、environment_id、status/type 等）非法值统一返回 400（不再静默忽略）
  - 关键枚举口径对齐：`/users?status=enabled|disabled`、`/resources?type=host|service`（不接受历史/错误枚举值）
- 前后端职责收口：
  - `environment_source_label` 这类展示文案从后端移除，仅保留稳定枚举值 `environment_source`
- gorm/gen 生成收口：
  - 生成入口收敛为 `go run ./cmd/gen-query`（不再并存 `cmd/gen-dao`）
  - model 托管试点扩展到 `environments/locations/hosts/services`（`internal/model/*.gen.go`），并剥离生成 model 的 `json tag`
- 仓储收缩与红线治理：
  - 删除/内联一批纯 CRUD 薄壳 repository（例如 user/credential/host_environment/service_environment/resource）
  - 正常请求链路不再访问 `information_schema`（运行时 schema 探测移除）
- Swagger：
  - 生成物：`backend/docs/swagger/*`
  - 访问：`/api/swagger/index.html`
- Page route_path：
  - canonical：自动剥离 `/admin` 前缀
  - CLI 迁移：`go run ./cmd/irms migrate-page-route-path`

## 兼容与迁移策略

- chi -> gin：保持 `/api/...` 前缀不变；Gin 作为唯一入口，legacy 仅保留为历史代码参考（不再承载运行时兜底）
- SQL -> gorm：新接口走 gorm repository；旧接口暂存（减少一次性切换风险）
- page route_path：对历史库提供迁移命令；对新增/同步入口自动规范化，避免继续产生脏数据

## 已验证项

- `go build ./...` 通过
- `go test ./...` 通过（含 MySQL 集成测试：environment_ids 清空与回退语义）
- Swagger 可打开（含 auth/users/user-groups/pages/resources/resource-groups/grants/credentials/permissions 等接口与模型展示）
- Host/Service 列表 VO 与 PUT 更新（environment_ids replace + 继承规则输出）可用

## 已知限制

- Raw SQL 仍有少量保留：原因与位置清单见 `docs/dev/RAW_SQL_NOTES.md`
