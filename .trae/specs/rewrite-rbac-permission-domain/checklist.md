# Checklist（修正版）

## 权限模型

- [x] `grants` 已使用 `permission_code`
- [x] `grants` 唯一键包含 `permission_code`
- [x] `permission_definitions` 为 grant 合法值真源
- [x] 不再使用统一 `ReadOnly / ReadWrite` 作为权限模型主表达
- [x] 本轮未引入 host -> service 自动授权蕴含

## 凭据权限

- [x] 仅保留 `host_credential.read / write`
- [x] 仅保留 `service_credential.read / write`
- [x] 已删除 `*_credential.reveal` 权限码
- [x] reveal 接口改为校验 `*_credential.read`
- [x] reveal 接口仍保留独立 endpoint 与审计日志
- [x] host credential 动作要求 `host.read` 前置
- [x] service credential 动作要求 `service.read` 前置

## Group 语义

- [x] 对外主参数统一为 `group_type / group_id`
- [x] 对外主字段统一为 `member_id / member_type / member_name`
- [x] 新响应主字段不再暴露 `resource_group_id / resource_key`
- [x] 新审计不再使用 `resource_group` 作为正式业务语义

## resources 兼容层

- [x] `resources` 不再作为正式授权主链路
- [x] `resource_groups` 不再作为正式对外领域模型
- [x] `ResourceController / ResourceService / resource_filter.go` 已标注 deprecated 或 compatibility
- [x] `/permissions/resources` 已改名或改语义
- [x] 不存在读路径补录 / legacy sync 恢复

## 文档一致性

- [x] spec / tasks / checklist 与正式项目文档一致
- [x] 正式项目文档已删除 `*_credential.reveal` 权限码描述
- [x] 权限管理页文档已改为 credential 只展示 `read / write`

## 工程验证

- [x] `DB_NAME must be irms` 继续成立
- [x] 未新增 `irms_*` 常驻开发库
- [x] `go run ./cmd/recreate-db` 通过
- [x] `go run ./cmd/gen-query` 通过
- [x] `go build ./...` 通过
- [x] `go test ./...` 通过
- [x] `swag init -g cmd/irms/main.go -o docs/swagger` 通过
- [x] `go run ./cmd/db-cleanup` 通过
