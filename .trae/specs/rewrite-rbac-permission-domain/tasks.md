# Tasks（修正版）

- [x] T1 事实基线确认
  - [x] 盘点 Grant / PermissionDefinition / Credential 权限检查 / ResourceGroup / Group Member / `/permissions/resources`
  - [x] 标注正式链路与 compatibility storage 边界
  - [x] 输出待修改文件清单

- [x] T2 凭据权限口径重写
  - [x] 从字典/schema/Swagger/controller/service 移除 `host_credential.reveal` / `service_credential.reveal`
  - [x] 统一凭据权限为 `host_credential.read|write`、`service_credential.read|write`
  - [x] reveal 接口改为校验 `*_credential.read`
  - [x] 保留 reveal 独立接口与审计日志
  - [x] 保留父对象最小读取前置条件

- [x] T3 Grant / PermissionDefinition 主链路收口
  - [x] 落地 grants 唯一键 `subject_type+subject_id+object_type+object_id+permission_code`
  - [x] 统一 `GrantRepository.Upsert / Update / Query` 逻辑
  - [x] 所有 grant 写入必须通过 `permission_definitions` 校验
  - [x] 收口 permission implication 实现方式
  - [x] 不重新引入 `ReadOnly / ReadWrite`

- [x] T4 Group 外部语义重写
  - [x] 对外契约统一为 `group_type/group_id/member_id/member_type/member_name`
  - [x] 旧字段仅兼容读取，不作为主输出字段
  - [x] Swagger/DTO/VO/controller/错误消息统一收口
  - [x] 审计 action/target_type/target_name_snapshot 统一为 `host_group/service_group`

- [x] T5 resources compatibility 降级
  - [x] 明确 `resources/resource_groups/resource_group_members` 为 compatibility storage
  - [x] 禁止新增围绕 `resources` 的新功能
  - [x] `ResourceController/ResourceService/resource_filter.go` 保留时显式标注 deprecated/compatibility
  - [x] `/permissions/resources` 改名或改语义，避免“我的所有资源”误解
  - [x] 不恢复任何 legacy sync / 读路径补录

- [x] T6 正式文档同步
  - [x] 更新正式项目文档并删除 `*_credential.reveal` 权限码描述
  - [x] 更新权限管理页文档为 credential 仅展示 `read/write`
  - [x] 更新行为矩阵与审计说明
  - [x] 更新 resources/resource_groups 正式定位
  - [x] 确保 spec/tasks/checklist/正式项目文档一致

- [x] T7 工程验证
  - [x] `go run ./cmd/recreate-db`
  - [x] `go run ./cmd/gen-query`
  - [x] `go build ./...`
  - [x] `go test ./...`
  - [x] `swag init -g cmd/irms/main.go -o docs/swagger`
  - [x] `go run ./cmd/db-cleanup`
  - [x] 验证不产生新的 `irms_*` 常驻开发库
  - [x] 结果记录：全部命令退出码为 0；`recreate-db` 输出 `database recreated: irms`；`db-cleanup` 输出 `PENDING_DELETE_DATABASES: <none>`

- [x] T8 交付报告
  - [x] 输出修改文件清单
  - [x] 输出 schema 变更说明
  - [x] 输出兼容字段保留与废弃说明
  - [x] 输出行为矩阵（credential、reveal、group 语义、`/permissions/resources` 新语义）
  - [x] 输出遗留兼容层清单

# Task Dependencies
- T2 depends on T1
- T3 depends on T1
- T4 depends on T1
- T5 depends on T1
- T6 depends on T2, T3, T4, T5
- T7 depends on T2, T3, T4, T5, T6
- T8 depends on T7
