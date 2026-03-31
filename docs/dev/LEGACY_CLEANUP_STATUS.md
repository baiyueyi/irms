# Legacy / Chi 清理状态

## 结论

- `/api` 运行时已无 legacy fallback（Gin 为唯一入口）。
- `backend/internal/httpapi/*`（chi 时代 legacy 实现）已移除。
- `github.com/go-chi/chi/v5` 已从 `backend/go.mod` 移除。

## 仍保留的“历史相关”内容

- `backend/cmd/recreate-db`：开发/验收用的辅助 CLI（非 legacy API 运行时依赖）。
- `docs/dev/RAW_SQL_NOTES.md`：Raw SQL 保留原因与迁移建议说明。

## 删除条件（下一轮）

- 若确认没有任何外部工具/文档仍引用已删除的 legacy httpapi 路由实现，则无需额外动作。
