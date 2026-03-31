# IRMS RBAC MVP 测试报告

## 测试环境

- OS: Windows（PowerShell 5）
- 前端: Vite + Vue3 + Pinia + Element Plus
- 后端: Go（在 `backend/` 目录执行 `go run .\cmd\irms`）
- 数据库: MySQL（使用现有环境变量连接）
- API 前缀: `/api/...`

## 环境变量使用说明（仅名称）

- `DB_HOST`
- `DB_NAME`
- `DB_PASSWORD`
- `DB_USER`
- `DB_PORT`
- `JWT_SECRET`
- `CREDENTIAL_ENCRYPTION_KEY`

## 手验范围

- A. SuperAdmin 初始化与重复初始化保护
- B. 登录与首次改密（超管 + 普通用户）
- C. 超管管理能力（用户/用户组/资源/资源组 CRUD + 组成员关系）
- D. Grants 与权限计算（四类授权、升级降级 UPDATE、去重、撤销即时生效）
- E. 普通用户侧访问边界与资源列表
- F. 审计日志动作覆盖与字段完整性

## 执行证据

- 联调证据文件：`docs/qa/_evidence_rbac_run.json`
  - run_tag: `20260317085806`
  - 结果：`all_pass = true`（11/11）
- CLI 初始化证据：
  - 首次初始化：`username: password: done`
  - 重复初始化：`panic: super_admin already exists`（符合预期，禁止重复）
- 浏览器手验证据（集成浏览器）：
  - 普通用户登录后 URL：`/my-resources`
  - 普通用户访问 `/admin/users` 后仍停留 `/my-resources`
  - 新建超管首次登录后 URL：`/change-password`

## 手验步骤与实际结果

| 编号 | 验证点 | 实际结果 |
|---|---|---|
| A1 | `init-superadmin` 首次执行 | 通过 |
| A2 | 已有超管再次执行 `init-superadmin` | 通过（阻止重复） |
| B1 | 超管首次登录需改密 | 通过（API 拦截 + 页面跳转 `/change-password`） |
| B2 | 超管改密后进入管理端 | 通过 |
| B3 | 普通用户首次登录需改密 | 通过 |
| B4 | 普通用户改密后可登录 | 通过 |
| C1 | 用户 CRUD | 通过 |
| C2 | 用户组 CRUD + 用户加入组 | 通过 |
| C3 | 资源 CRUD | 通过 |
| C4 | 资源组 CRUD | 通过 |
| C5 | 资源组同型校验 | 通过（返回 `RESOURCE_TYPE_MISMATCH`） |
| D1 | 用户直授资源 | 通过 |
| D2 | 用户组授资源 | 通过 |
| D3 | 用户授资源组 | 通过 |
| D4 | 用户组授资源组 | 通过 |
| D5 | ReadOnly/ReadWrite 生效 | 通过 |
| D6 | Grant 升级/降级走 UPDATE | 通过 |
| D7 | 不产生重复有效 grant | 通过（同键 total=1） |
| D8 | 撤销后权限立即变化 | 通过 |
| E1 | 普通用户不可访问管理 API | 通过（`FORBIDDEN`） |
| E2 | 普通用户仅留在“我的可访问资源”页 | 通过 |
| E3 | `GET /api/permissions/resources` 与授权一致 | 通过 |
| F1 | 关键动作写入 audit_logs | 通过 |
| F2 | 审计字段最小集合齐全 | 通过 |

## 失败项

- 首轮联调失败（已修复）：
  - `user_groups.description`、`resource_groups.type` 等历史库缺列导致 1054 错误
  - `resource_group_members.resource_id` 历史列约束导致插入失败
  - 修复后重跑通过

## 修复结果

- 后端迁移补齐历史库兼容：
  - 启动时自动补列（按 `information_schema` 检测）
  - 兼容旧表 `resource_id` 场景（插入时双写 `resource_key/resource_id`）
- 登录改密链路修复：
  - `must_change_password` 不再依赖 JWT claims，改为服务端按用户实时状态校验
  - 改密后沿用当前 token 访问受保护接口不再被 `FIRST_LOGIN_PASSWORD_CHANGE_REQUIRED` 拦截
- 重跑端到端后：`all_pass = true`

## 未完成项

- 无阻塞验收的未完成项

## 已知限制

- `audit_logs.ip` 在部分本地代理链路下可能记录为非标准值（如 `[`）；字段存在且可空，后续可单独增强 IP 解析策略。
- 当前页面为 MVP 验收形态，已补最小编辑/删除/筛选与提示，未扩展复杂批量操作与高级筛选。
