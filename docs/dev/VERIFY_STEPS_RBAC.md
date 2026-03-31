# IRMS RBAC MVP PowerShell 验证步骤

## 0. 前置条件

- 已在当前会话中配置环境变量（仅示例名称）：`DB_HOST` `DB_PORT` `DB_NAME` `DB_USER` `DB_PASSWORD` `JWT_SECRET` `CREDENTIAL_ENCRYPTION_KEY`
- 不使用 dotenv

## 1. 后端启动

```powershell
cd d:\Projects\irms\backend
go run .\cmd\irms
```

期望输出：

- `server listening on :8080`

后端构建检查：

```powershell
cd d:\Projects\irms\backend
go build ./...
```

## 2. 前端启动

```powershell
cd d:\Projects\irms\frontend
npm run dev -- --host 0.0.0.0 --port 5173
```

期望输出：

- `Local: http://localhost:5173/`（若端口占用会自动切到 5174）

## 3. SuperAdmin 初始化

```powershell
cd d:\Projects\irms\backend
$u='sa_e2e'
$p='Init#12345'
"$u`n$p`n" | go run .\cmd\irms init-superadmin
```

重复执行验证：

```powershell
$u='sa_retry'
$p='Init#999'
"$u`n$p`n" | go run .\cmd\irms init-superadmin
```

期望：第二次直接失败并提示已存在超管。

## 4. 登录与首次改密

- 浏览器访问：`http://localhost:5173/login`
- 超管首次登录应跳转：`/change-password`
- 完成改密后进入管理端，后续点击管理操作不应再出现 `password change required`
- 普通用户首次登录同样先改密

## 5. 授权与撤销（API 最小验证）

```powershell
$base='http://localhost:8080/api'
$login = Invoke-RestMethod -Method Post -Uri "$base/auth/login" -ContentType 'application/json' -Body (@{username='sa_e2e';password='Init#12345!N'}|ConvertTo-Json)
$token = $login.data.token
$h=@{Authorization="Bearer $token"}

# 示例：创建授权（若已存在同唯一键则 UPDATE）
Invoke-RestMethod -Method Post -Uri "$base/grants" -Headers $h -ContentType 'application/json' -Body (@{
  subject_type='user';subject_id=1;object_type='resource';object_id=1;permission='ReadOnly'
}|ConvertTo-Json)
```

撤销授权：

```powershell
Invoke-RestMethod -Method Delete -Uri "$base/grants/{grantId}" -Headers $h
```

## 6. 普通用户验证

- 普通用户登录后访问：
  - `GET /api/users` => 应返回 `FORBIDDEN`
  - `GET /api/permissions/resources` => 返回可访问资源列表
- 页面验证：
  - 普通用户进入 `/admin/users` 时应被路由守卫重定向到 `/my-resources`

## 7. 一键端到端证据复跑

已在本地落盘执行证据文件：

- `docs/qa/_evidence_rbac_run.json`

可用其作为验收基准（包含 CRUD、同型校验、授权计算、撤销即时生效、审计覆盖结果）。
