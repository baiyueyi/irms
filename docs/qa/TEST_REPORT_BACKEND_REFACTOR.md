# 后端重构验收报告（Gin + Gorm + Swagger）

## 范围

- A. gin + gorm 启动链路与路由生效
- B. Swagger 生成与可访问性
- C. Host/Service 列表 VO 输出（环境信息与来源规则）
- D. Host/Service 统一 PUT 更新 + environment_ids 一次性 replace
- E. page route_path canonical 规则与迁移命令可执行

## 环境

- OS：Windows
- Workspace：`D:\Projects\irms`
- 后端：`backend/`
- 说明：为避免影响既有数据，本报告使用独立 DB：`DB_NAME=irms_refactor_test`

## 执行步骤与结果

### A1. go build 通过

```powershell
cd d:\Projects\irms\backend
go build ./...
```

- 结果：通过
- 说明：`internal/service` 下部分集成测试依赖 MySQL；当未配置或 MySQL 不可达时会自动 Skip（不再导致整包失败）。

### A2. 使用独立数据库启动并初始化超管

```powershell
cd d:\Projects\irms\backend
$env:DB_NAME='irms_refactor_test'
go run .\cmd\recreate-db

$u='sa_refactor'
$p='Init#12345'
"$u`n$p`n" | go run .\cmd\irms init-superadmin
```

- 结果：通过（数据库创建成功；超管初始化成功）

### A3. 启动服务

```powershell
cd d:\Projects\irms\backend
$env:DB_NAME='irms_refactor_test'
$env:SERVER_ADDR=':8081'
go run .\cmd\irms
```

- 结果：通过（监听 `:8081`）

### B1. Swagger 生成

见 [SWAGGER_BACKEND.md](file:///d:/Projects/irms/docs/dev/SWAGGER_BACKEND.md)

### B2. Swagger 可访问

- 访问：`http://127.0.0.1:8081/api/swagger/index.html`
- 结果：可打开，并展示 auth/users/user-groups/pages/resource-groups/grants/credentials/permissions 等接口与模型
- 关键截图：`docs/qa/evidence_backend_swagger_fulltags.png`

### C1. 登录与改密（首次登录）

```powershell
$base='http://127.0.0.1:8081/api'
$login = Invoke-RestMethod -Method Post -Uri "$base/auth/login" -ContentType 'application/json' -Body (@{username='sa_refactor';password='Init#12345'}|ConvertTo-Json)
$token = $login.data.token
$h=@{Authorization="Bearer $token"}

Invoke-RestMethod -Method Post -Uri "$base/auth/change-password" -Headers $h -ContentType 'application/json' -Body (@{old_password='Init#12345';new_password='Init#12345!N'}|ConvertTo-Json)
```

- 结果：通过（改密成功，后续可访问受保护接口）

### D1. 构造数据（复用 legacy API 创建）

```powershell
$base='http://127.0.0.1:8081/api'
$token=(Invoke-RestMethod -Method Post -Uri "$base/auth/login" -ContentType 'application/json' -Body (@{username='sa_refactor';password='Init#12345!N'}|ConvertTo-Json)).data.token
$h=@{Authorization="Bearer $token"}

$env1=(Invoke-RestMethod -Method Post -Uri "$base/environments" -Headers $h -ContentType 'application/json' -Body (@{code='dev';name='Dev';status='active'}|ConvertTo-Json)).data.id
$env2=(Invoke-RestMethod -Method Post -Uri "$base/environments" -Headers $h -ContentType 'application/json' -Body (@{code='prod';name='Prod';status='active'}|ConvertTo-Json)).data.id
$loc=(Invoke-RestMethod -Method Post -Uri "$base/locations" -Headers $h -ContentType 'application/json' -Body (@{code='bj';name='Beijing';location_type='idc';status='active'}|ConvertTo-Json)).data.id

$hostId=(Invoke-RestMethod -Method Post -Uri "$base/hosts" -Headers $h -ContentType 'application/json' -Body (@{name='host-a';hostname='host-a.internal';primary_address='10.0.0.1';provider_kind='vm';os_type='linux';status='active';location_id=$loc}|ConvertTo-Json)).data.id
Invoke-RestMethod -Method Post -Uri "$base/host-environments" -Headers $h -ContentType 'application/json' -Body (@{host_id=$hostId;environment_id=$env1}|ConvertTo-Json) | Out-Null
Invoke-RestMethod -Method Post -Uri "$base/host-environments" -Headers $h -ContentType 'application/json' -Body (@{host_id=$hostId;environment_id=$env2}|ConvertTo-Json) | Out-Null

$svcId=(Invoke-RestMethod -Method Post -Uri "$base/services" -Headers $h -ContentType 'application/json' -Body (@{name='svc-a';service_kind='app';host_id=$hostId;endpoint_or_identifier='svc-a.internal';port=8080;protocol='http';status='active'}|ConvertTo-Json)).data.id
```

- 结果：通过

### D2. 验证 HostVO（含环境 codes/names）

```powershell
Invoke-RestMethod -Method Get -Uri "$base/hosts" -Headers $h | ConvertTo-Json -Depth 8
```

- 结果：通过（environment_ids/codes/names 均由后端输出）

### D3. 验证 ServiceVO（继承规则由后端输出）

```powershell
Invoke-RestMethod -Method Get -Uri "$base/services" -Headers $h | ConvertTo-Json -Depth 10
```

- 结果：通过（service 无自身环境时 `environment_source=host_inherited`，effective 环境等于 host 环境）

### D4. 验证统一 PUT 更新（environment_ids 一次性 replace）

```powershell
# 让 service 拥有自身环境（覆盖 host 继承）
Invoke-RestMethod -Method Put -Uri "$base/services/$svcId" -Headers $h -ContentType 'application/json' -Body (@{
  name='svc-a';service_kind='app';host_id=$hostId;endpoint_or_identifier='svc-a.internal';port=8080;protocol='http';status='active';
  environment_ids=@($env2)
}|ConvertTo-Json) | ConvertTo-Json -Depth 10

Invoke-RestMethod -Method Get -Uri "$base/services" -Headers $h | ConvertTo-Json -Depth 10
```

- 结果：通过（update 返回最新 VO；list 中 `environment_source=own` 且 effective 环境为自身环境）

### E1. route_path canonical 与迁移命令

- canonical：`/admin/...` 会被自动规范化为 `/...`（create/update/sync）
- 迁移命令：

```powershell
cd d:\Projects\irms\backend
go run .\cmd\irms migrate-page-route-path
```

- 结果：命令可执行；用于批量修复历史库中已存的 `/admin/...` 数据

## 本轮收口（2026-03-24）

### R1. go build / go.mod

```powershell
cd d:\Projects\irms\backend
go version
go mod tidy
go build ./...
```

- 结果：通过（本机 Go：`go1.25.6`；`go build ./...` 通过）

### R2. Swagger 生成与访问

```powershell
cd d:\Projects\irms\backend
go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/irms/main.go -o .\docs\swagger --parseInternal --dir .
```

- 结果：通过（`backend/docs/swagger/swagger.json` 更新成功）
- 访问：`http://127.0.0.1:8081/api/swagger/index.html`

### R3. 使用独立 DB 验证核心链路

```powershell
cd d:\Projects\irms\backend
$env:DB_NAME='irms_refactor_verify3'
go run .\cmd\recreate-db

$u='sa_refactor'
$p='Init#12345'
"$u`n$p`n" | go run .\cmd\irms init-superadmin

$env:SERVER_ADDR=':8081'
go run .\cmd\irms
```

### R4. Host / Service typed PUT 更新（含 environment_ids replace）

关键断言：

- `PUT /api/hosts/:id`、`PUT /api/services/:id` 入参为 typed request（不再 RawMessage/patch）。
- `environment_ids` 作为主更新对象字段提交，由 service 层在事务内 replace 绑定表。
- `GET /api/hosts/:id`、`GET /api/services/:id` 返回 VO 中包含 own/inherited/effective/source。

（完整联调脚本见终端日志；关键结果示例）

- `GET /api/hosts/:id`：`own_environment_ids=[1,2]`、`environment_source=host`
- `GET /api/services/:id`：`own_environment_ids=[2]`、`inherited_environment_ids=[1,2]`、`environment_ids=[2]`、`environment_source=service`

### R5. /hosts/:id/services 已迁移到 Gin

- `GET /api/hosts/:id/services?page=1&page_size=20` 返回分页结构 `data.list + data.pagination`，并且字段与 `GET /api/services` 输出一致。

### R6. 审计接入新链路

- `GET /api/audit-logs?page=1&page_size=20` 可返回新链路写操作审计记录。
- 覆盖样例（本轮联调链路中已出现）：`create_host`、`update_host`、`replace_host_environments`、`create_service`、`update_service`、`replace_service_environments`、`create_host_credential`、`reveal_credential`。

### R7. reveal credential typed response

- `POST /api/host-credentials/:id/reveal` 返回 typed VO（示例：`{id, secret}`）。

### R8. legacy fallback 收口

- Gin 路由不再使用 `NoRoute -> legacy.ServeHTTP` 兜底；`/api` 的未注册路径将直接返回 404。

## 本轮边角收口（2026-03-25）

### S1. environment_ids 支持显式清空

- Host update：`environment_ids: []` 合法，语义为 replace 为 0 条绑定
- Service update：`environment_ids: []` 合法，语义为清空自身环境，生效环境回退为继承 host（若 host 也无环境则为无环境）
- 通过 MySQL 集成测试覆盖：
  - `TestHostUpdateEnvironmentIDsClear`
  - `TestServiceUpdateEnvironmentIDsClearFallbackToHost`

### S2. Host / Service 创建接口支持 environment_ids 单对象提交

- `POST /api/hosts`、`POST /api/services` 请求体必须包含 `environment_ids` 字段（可为 `[]`），创建后在同一事务中 replace 绑定表。

### S3. change-password 补齐审计

- `POST /api/auth/change-password`：成功/失败均写入 `audit_logs`（action=`change_password`），不记录任何密码明文。

### S4. 关键写接口启用严格 JSON 绑定

- 对未知 JSON 字段返回 400（不再静默吞掉拼错/旧字段）。
- 示例：`POST /api/hosts` 携带 `unknown_field` 返回 400。

### S5. legacy / chi 收口

- `backend/internal/httpapi/*` 已移除，`github.com/go-chi/chi/v5` 已从 go.mod 移除。

### S6. Raw SQL 边界收敛

- pages/grants 等 Raw SQL 已集中到 repository 层；说明文档更新：`docs/dev/RAW_SQL_NOTES.md`。

## 本轮边界收口（2026-03-26）

### T1. go build / go test

```powershell
cd d:\Projects\irms\backend
go build ./...
go test ./...
```

- 结果：通过

Swagger 生成：

```powershell
cd d:\Projects\irms\backend
go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/irms/main.go -o .\docs\swagger --parseInternal --dir .
```

- 结果：通过（swagger.json/yaml/docs.go 更新）

### T5. 后端移除展示 label（environment_source_label）

- `vo.HostVO` / `vo.ServiceVO` 已移除 `environment_source_label` 输出字段，仅保留稳定枚举值 `environment_source`。

### T2. 列表接口 query 参数非法值 -> 400（不再静默忽略）

```powershell
cd d:\Projects\irms\backend
$env:DB_NAME='irms_refactor_verify_query400'
go run .\cmd\recreate-db

$u='sa_refactor'
$p='Init#12345'
"$u`n$p`n" | go run .\cmd\irms init-superadmin

$env:SERVER_ADDR=':8082'
go run .\cmd\irms
```

关键断言（示例）：

- `GET /api/hosts?environment_id=abc`：

```json
{"code":"INVALID_ARGUMENT","message":"invalid query","details":{"errors":[{"field":"environment_id","tag":"invalid"}]}}
```

- `GET /api/host-environments`（缺少必填 query `host_id`）：

```json
{"code":"INVALID_ARGUMENT","message":"invalid query","details":{"errors":[{"field":"host_id","tag":"required"}]}}
```

- 覆盖校验点：`page`、`page_size`、`environment_id`、`status`、`type`、`resource_group_id`、`host_id/service_id` 等在关键列表接口中均已改为非法直接 400。

补充断言（枚举契约对齐）：

- `GET /api/users?status=enabled` -> 200；`GET /api/users?status=active` -> 400：

```json
{"code":"INVALID_ARGUMENT","message":"invalid query","details":{"errors":[{"field":"status","tag":"invalid"}]}}
```

- `GET /api/resources?type=host` -> 200；`GET /api/resources?type=page` -> 400：

```json
{"code":"INVALID_ARGUMENT","message":"invalid query","details":{"errors":[{"field":"type","tag":"invalid"}]}}
```

- `GET /api/hosts?provider_kind=xxx` -> 400：

```json
{"code":"INVALID_ARGUMENT","message":"invalid query","details":{"errors":[{"field":"provider_kind","tag":"invalid"}]}}
```

- `GET /api/services?service_kind=xxx` -> 400：

```json
{"code":"INVALID_ARGUMENT","message":"invalid query","details":{"errors":[{"field":"service_kind","tag":"invalid"}]}}
```

写接口一致性检查（示例）：

- `POST /api/users` 携带 `status=active` -> 400（validator oneof）：

```json
{"code":"INVALID_ARGUMENT","message":"invalid request","details":{"errors":[{"field":"status","tag":"oneof","param":"enabled disabled"}]}}
```

- `POST /api/resources` 携带 `type=page` -> 400（validator oneof）：

```json
{"code":"INVALID_ARGUMENT","message":"invalid request","details":{"errors":[{"field":"type","tag":"oneof","param":"host service"}]}}
```

### T3. gorm/gen 生成入口收口 + model 托管扩展（environments/locations/hosts/services）

生成器（唯一入口）：

```powershell
cd d:\Projects\irms\backend
$env:DB_NAME='irms_refactor_gen'
go run .\cmd\recreate-db
go run .\cmd\gen-query
```

- 结果：通过
- 断言：
  - `internal/model` 中 `environments/locations/hosts/services` 均由 `*.gen.go` 托管，且已剥离 `json tag`（仅保留 gorm tag）
  - `cmd/gen-dao` 已删除（不再并存两套生成入口）

### T4. Swagger 与 400 契约同步（示例核对）

- `GET /api/users` 与 `GET /api/resources` 等列表接口 swagger 已包含：
  - query 参数枚举（Enums）
  - 400 response schema（`controller.ErrorResponse`）
