# Raw SQL 保留清单（临时）

说明：当前后端已切到 Gin + Gorm，但以下场景仍保留 Raw SQL（或 gorm.Exec 直接 SQL）以保证查询表达清晰、联调稳定、兼容历史 schema。

## 仍保留的位置与原因

- pages 列表：需要输出 `grant_count`（子查询计数），且要同时支持关键字/状态过滤与分页。
  - 位置：`internal/repository/page_repository.go`
- grants 授权校验与展示：需要做 “user 直接授权 + user_group 继承授权” 的计数/判定，以及若干对象的名称快照查询。
  - 位置：`internal/repository/grant_repository.go`
- resource-group-members 列表：需要 `LEFT JOIN resources` 做 name 关键字过滤与分页。
  - 位置：`internal/repository/resource_group_member_repository.go`
- resource-group-members 兼容：历史库可能存在 `resource_id` 字段差异，需要通过 `information_schema.columns` 探测后选择插入语句。
  - 位置：`internal/repository/resource_group_member_repository.go`
- resource-groups 列表：需要输出 `member_count / grant_count`（聚合计数），并支持过滤与分页。
  - 位置：`internal/repository/resource_group_repository.go`
- permissions（我的可访问资源）：需要做 “user 直接授权 + user_group 继承授权” 的聚合与 ReadWrite 优先规则。
  - 位置：`internal/repository/permission_repository.go`

## 迁移建议（下一轮）

- 对 “计数/聚合字段” 的查询：优先用 gorm 的 `Select + Joins + Group` 重写；若可读性下降明显，则考虑保留 Raw 并配套测试覆盖。
- 对 “schema 探测/兼容” 的逻辑：待线上库统一后，可移除 `information_schema` 探测分支，收敛为单一写路径。
