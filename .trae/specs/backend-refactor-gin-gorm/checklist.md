# Checklist

## 架构

- [x] gin 为唯一主 HTTP 框架（Router、middleware、分组清晰）
- [x] gorm 为主要数据访问方式（controller 不直接依赖 DB）
- [x] 分层目录齐全：model/dto/vo/repository/service/controller/middleware/pkg

## API 语义

- [x] 响应形状统一（code/message/data/request_id/details）
- [x] 错误码稳定且不透传内部 err.Error()
- [x] 参数绑定统一使用 gin binding/validator

## editDraft 支撑

- [x] `PUT /api/hosts/:id` 支持 typed request + environment_ids 一次性 replace
- [x] `PUT /api/services/:id` 支持 typed request + environment_ids 一次性 replace
- [x] 列表接口返回 VO（包含 environments 与 source 规则）

## Page route_path

- [x] route_path 视为 canonical path
- [x] 提供 `/admin/...` -> `/...` 的迁移方案并可执行

## Swagger

- [x] swagger 可生成（给出 PowerShell 步骤）
- [x] swagger 路由可打开并已手验

## 兼容与迁移

- [x] legacy 接口可临时访问（避免前端/已有数据瞬间失效）
- [x] 给出迁移执行步骤与风险点

## 验证

- [x] `go build ./...` 通过
- [x] 服务可运行
- [x] 核心接口验证通过（至少 auth/hosts/services/environments/pages）
