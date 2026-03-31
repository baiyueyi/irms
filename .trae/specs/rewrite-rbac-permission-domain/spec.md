# IRMS Backend 局部重写：RBAC / Permission Domain Spec

## Why
当前权限域同时存在新旧模型混用：权限码、group 语义、兼容层暴露和接口命名均不一致，导致运行时行为与文档割裂。需要在现有工程内局部重写权限域，形成可长期维护的统一后端语义。

## What Changes
- 重写 Grant / PermissionDefinition / Permission Check 主链路，统一由 `permission_definitions` 驱动。
- 凭据权限收口为 `read/write` 两档，移除所有 `*_credential.reveal` 权限码。
- reveal 保留为独立接口动作，改为校验 `*_credential.read`，且必须审计。
- 保留 credential 父对象 read 前置：host credential 需 `host.read`，service credential 需 `service.read`。
- 保留 service 绑定 host 的前置链；本轮不引入 host 自动蕴含 service 权限。
- grants 唯一键改为 `(subject_type, subject_id, object_type, object_id, permission_code)`。
- 对外 group 语义统一为 `host_group/service_group` 与 `group_*` / `member_*` 字段。
- `resources/resource_groups/resource_group_members` 降级为 compatibility storage，不再作为正式主链路概念。
- `/permissions/resources` 完成改名或语义收口，避免“资源泛化”误导。
- `ResourceController/ResourceService/resource_filter.go` 保留时必须明确 deprecated/compatibility。
- 继续硬化单库约束：运行期与工具链仅允许 `DB_NAME=irms`。

## Impact
- Affected specs: RBAC 授权模型、权限字典、凭据权限、group 兼容层、权限接口命名、审计语义、单库约束。
- Affected code:
  - `internal/db/schema.go`
  - `internal/service/grant_service.go`
  - `internal/repository/grant_repository.go`
  - `internal/service/permission_definition_service.go`
  - `internal/controller/grant_controller.go`
  - `internal/service/credential_service.go`
  - `internal/service/resource_group_service.go`
  - `internal/controller/resource_group_controller.go`
  - `internal/dto/request/resource_group*.go`
  - `internal/vo/resource_group*.go`
  - `internal/repository/permission_repository.go`
  - `internal/controller/permission_controller.go`
  - swagger 与正式项目文档

## ADDED Requirements
### Requirement: 权限字典真源
系统 SHALL 将 `permission_definitions` 作为 `permission_code` 的唯一合法来源。

#### Scenario: grant 写入校验
- **WHEN** 创建或更新 grant
- **THEN** 必须基于 `permission_definitions` 校验 `permission_code` 合法性
- **AND** 非法值拒绝写入

### Requirement: 凭据 reveal 动作与权限分离
系统 SHALL 将 reveal 作为接口动作而非权限码。

#### Scenario: reveal 访问
- **WHEN** 访问 host/service 凭据 reveal 接口
- **THEN** 需要 `*_credential.read`
- **AND** 需要对应父对象 `*.read` 前置
- **AND** 必须记录审计日志

### Requirement: grants 多动作共存
系统 SHALL 允许同一主体对同一对象拥有多条不同 `permission_code` grant。

#### Scenario: 同对象多授权
- **WHEN** 同一主体对同一对象分别授予 read 和 write
- **THEN** 两条记录并存，不互相覆盖

## MODIFIED Requirements
### Requirement: Group 对外契约
系统 SHALL 对外统一使用 `host_group/service_group` 与 `group_type/group_id/member_id/member_name/member_type`。旧 `resource_*` 仅兼容读取，禁止作为主字段。

### Requirement: resources 兼容层定位
系统 SHALL 将 `resources/resource_groups/resource_group_members` 明确为 compatibility storage 与 migration bridge，不再作为正式 RBAC 主概念。

### Requirement: 权限蕴含策略
系统 SHALL 在运行时仅实现 `write -> read` 判定展开，不写回 grants；本轮不实现 host->service 自动授权蕴含。

## REMOVED Requirements
### Requirement: `*_credential.reveal` 作为权限码
**Reason**: reveal 是接口动作，不应作为权限码，避免模型重复。  
**Migration**: 删除 reveal 码在字典/接口枚举/校验中的所有暴露，reveal 接口统一校验 `*_credential.read`。

### Requirement: `resource_group` 作为正式业务语义
**Reason**: 正式业务语义已切换至 `host_group/service_group`。  
**Migration**: 旧表仅内部承载；controller/dto/vo/swagger/audit 全部切换新语义。
