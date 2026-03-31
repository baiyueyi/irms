# Trae Rules（IRMS 后端代码规范）

1. **gorm/gen 优先**：数据库映射 Model 尽量由 `gorm/gen` 管理；表字段变更后优先通过生成更新 Model，不手工长期维护数据库映射字段。
2. **分层明确**：`model` 只做数据库映射；`dto/request` 只做入参；`dto/response` / `vo` 只做出参展示。禁止一个 struct 同时承担 Model + DTO + VO。
3. **Model 不承载输出语义**：Model 中尽量不要放 `json` tag、展示字段、聚合字段、前端专用字段；输出统一走 `dto/response` 或 `vo`。
4. **Controller 职责**：只做参数绑定、校验、调用 service、返回统一响应；禁止在 controller 写 SQL、拼复杂业务、直接返回 model。
5. **Service 职责**：负责业务编排、事务、权限、审计；可直接使用 `gorm/gen` 基础查询；不要把简单 CRUD 全部机械下沉成空心 repository。
6. **Repository 边界**：只有复杂聚合、复杂搜索、批量 replace、权限判定、性能敏感查询、Raw SQL 等场景才封装 repository；`GetByID/ListByIDs/Create/Delete` 等若 gen 已满足，直接用。
7. **禁止过度封装**：不要为了“有 repository”而包装一层无业务价值的方法；优先减少空壳代码。
8. **连接池单例**：全局只允许一套底层连接池；可同时持有 `*gorm.DB`、`*sql.DB`、`*query.Query`，但三者必须共用同一底层连接。
9. **事务规范**：事务内必须使用事务态 `tx` 和 `query.Use(tx)`；禁止事务里一半走 `tx`、一半走全局 query；事务态对象不要暴露误导性的全局 SQL 指针。
10. **Raw SQL 约束**：优先用 gorm/gen；仅在复杂查询、聚合、权限判定、性能必要时使用 Raw SQL；Raw SQL 优先放 repository/query 层，不放 controller；保留 Raw SQL 时写明原因。
11. **查询优化**：消灭明显 N+1；列表查询优先批量查、join、map 回填；不要在循环里逐条补名称/环境/统计。
12. **接口契约**：写接口严格绑定并拒绝非法字段；列表 query 参数非法值要明确返回 400，不要静默忽略。
13. **Swagger 同步**：新增/修改接口时同步更新 Swagger 注释；Swagger 必须对应真实请求 DTO 和响应 VO。
14. **代码取向**：优先“边界清晰、可维护、可验证”，不要为了形式上的优雅新增抽象层。
