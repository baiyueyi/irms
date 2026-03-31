# 后端接口迁移清单（Gin vs Legacy）

说明：

- 所有路径均以 `/api` 为前缀（下文省略 `/api`）。
- “已迁移”指：Gin 已注册对应路由，实际请求不会落到 legacy fallback。

## 已迁移（Gin）

**公共**

- GET `/health`
- GET `/swagger/*any`

**auth**

- POST `/auth/login`
- POST `/auth/change-password`
- GET `/me`

**hosts**

- GET `/hosts`
- GET `/hosts/:id`
- POST `/hosts`
- PUT `/hosts/:id`
- DELETE `/hosts/:id`
- GET `/hosts/:id/services`

**services**

- GET `/services`
- GET `/services/:id`
- POST `/services`
- PUT `/services/:id`
- DELETE `/services/:id`

**environments**

- GET `/environments`
- POST `/environments`
- PUT `/environments/:id`
- DELETE `/environments/:id`

**locations**

- GET `/locations`
- POST `/locations`
- PUT `/locations/:id`
- DELETE `/locations/:id`

**host-environments / service-environments**

- GET `/host-environments`
- POST `/host-environments`
- DELETE `/host-environments`
- GET `/service-environments`
- POST `/service-environments`
- DELETE `/service-environments`

**users**

- GET `/users`
- POST `/users`
- PUT `/users/:id`
- DELETE `/users/:id`

**user-groups**

- GET `/user-groups`
- POST `/user-groups`
- PUT `/user-groups/:id`
- DELETE `/user-groups/:id`
- GET `/user-group-members`
- POST `/user-group-members`
- DELETE `/user-group-members`

**pages**

- GET `/pages`
- POST `/pages`
- PUT `/pages/:id`
- DELETE `/pages/:id`
- POST `/pages/sync`

**resources**

- GET `/resources`
- POST `/resources`
- PUT `/resources/:key`
- DELETE `/resources/:key`

**resource-groups**

- GET `/resource-groups`
- POST `/resource-groups`
- PUT `/resource-groups/:id`
- DELETE `/resource-groups/:id`
- GET `/resource-group-members`
- POST `/resource-group-members`
- DELETE `/resource-group-members`

**grants**

- GET `/grants`
- POST `/grants`
- PUT `/grants/:id`
- DELETE `/grants/:id`

**credentials**

- GET `/host-credentials`
- POST `/host-credentials`
- PUT `/host-credentials/:id`
- DELETE `/host-credentials/:id`
- POST `/host-credentials/:id/reveal`
- GET `/service-credentials`
- POST `/service-credentials`
- PUT `/service-credentials/:id`
- DELETE `/service-credentials/:id`
- POST `/service-credentials/:id/reveal`

**permissions**

- GET `/permissions/resources`

**audit-logs**

- GET `/audit-logs`

## 仍在 legacy fallback（Chi / database/sql）

- 无（所有 `/api` 接口已迁移到 Gin）

## 下一轮迁移顺序（建议）

1. 移除 legacy httpapi 代码依赖：逐步下线 `internal/httpapi/*`，保留到必要迁移完成后再删除目录
2. 收敛 Raw SQL：按“仍保留 Raw SQL 清单”逐步替换为 gorm 查询或视图/存储过程（仅在必要时）
