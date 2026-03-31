# Host / Service 的 environment_ids 契约

## 适用接口

- Host
  - `POST /api/hosts`
  - `PUT /api/hosts/:id`
- Service
  - `POST /api/services`
  - `PUT /api/services/:id`

## 契约规则（强约束）

- 请求体必须包含 `environment_ids` 字段
  - 缺失该字段：返回 400（invalid request）
- `environment_ids: []` 合法
  - Host：表示清空主机环境绑定
  - Service：表示清空服务自身环境绑定，生效环境将回退为继承 Host（若 Host 也无环境则为无环境）
- `environment_ids: [1,2,...]` 合法
  - 语义为 replace（以该数组为准覆盖绑定表）

## 前端对接建议

- editDraft 初始化：
  - Host：使用 `GET /api/hosts/:id` 的 `own_environment_ids` 回填编辑态
  - Service：使用 `GET /api/services/:id` 的 `own_environment_ids` 回填编辑态
- 提交：
  - create/update 都显式传 `environment_ids`（即使为空也传 `[]`），避免“漏传字段导致 400”
