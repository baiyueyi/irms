# 数据访问层规范（Gorm + gorm/gen）

## 目标

- 主查询方式统一为 gorm/gen（`internal/query`）。
- repository 只保留“有业务意义的复杂数据访问”（聚合/Join/权限判定/批量 replace/必须 Raw SQL）。
- 同时保留原生 SQL 能力，但 Gorm/SQL 必须共用同一个连接池。

## 连接池与三个指针

启动时通过 `Store` 同时初始化：

- `Store.Gorm`：`*gorm.DB`（GORM 主入口）
- `Store.SQL`：`*sql.DB`（原生 SQL 指针）
- `Store.Query`：`*query.Query`（gorm/gen 查询入口）

约束：

- `*sql.DB` 只创建一次（`db.OpenMySQL`）。
- `*gorm.DB` 使用同一个 `*sql.DB` 作为底层连接（`bootstrap.OpenGormFromSQLConn`）。
- `Store.New` 会校验 `gormDB.DB()` 与传入 `*sql.DB` 指针一致，避免出现第二套连接池。

实现位置：

- `backend/internal/store/store.go`
- `backend/cmd/irms/main.go`

## gorm/gen Query 初始化

- 生成入口：
  - `backend/cmd/gen-query`：生成试点表的 `backend/internal/model/*.gen.go` + `backend/internal/query/*`（统一入口，生成物禁止手改）。
- 运行时初始化：`Store.New` 调用 `query.SetDefault(gormDB)`，并在 `Store.Query` 中暴露统一入口。

### 本轮 model 托管试点

- 已切换为 gorm/gen 托管（生成 `*.gen.go`，禁止手改）：
  - `environments` → `internal/model/environments.gen.go`
  - `locations` → `internal/model/locations.gen.go`
  - `hosts` → `internal/model/hosts.gen.go`
  - `services` → `internal/model/services.gen.go`

### model 里的 json tag 策略

- gorm/gen 默认会为 model 生成 `json` tag，但 model 层不承载输出语义。
- `cmd/gen-query` 会在生成结束后对 `internal/model/*.gen.go` 做一次自动处理：剥离 `json:"..."` tag，确保 model 只保留数据库映射相关 tag。

## 事务中如何使用 Query

禁止在事务中混用“全局 Query + tx”。

推荐：

- 使用 `Store.Gorm.Transaction(func(tx *gorm.DB) error { ... })` 时：
  - 事务内查询入口用 `query.Use(tx)` 或 `Store.WithTx(tx).Query`（事务态仅暴露 Gorm+Query，不暴露 SQL）
- 或直接使用 `Store.Query.Transaction(func(tx *query.Query) error { ... })`（由 gen 提供）

原生 SQL 在事务中的正确姿势：

- 用 `tx.Exec(...)` / `tx.Raw(...)`（Gorm 会把 SQL 放在同一个事务里执行）
- 不要使用 `Store.SQL.Exec(...)`（它永远是全局连接池句柄，不是事务句柄）

## repository 边界

允许保留 repository 的场景：

- 复杂聚合/Join/Union/GroupBy
- 权限判定（如 grants/permissions）
- 复杂分页与搜索
- 批量 replace（如 host/service environments 绑定替换）
- 明显需要 Raw SQL 的场景

不建议保留的场景：

- 简单 CRUD（Create/Update/Delete）
- 简单按 ID 查询
- 简单条件查询（Where/First/Find）

这些场景优先直接用 `Store.Query`（gorm/gen 已生成基础能力）。
