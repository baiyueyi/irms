---
topic: rbac-mvp
status: draft
---

# Checklist

## Scope

- 仅实现：用户、用户组、page、host、service、environment、location、credential、授权关系、审计日志
- 仅实现角色：SuperAdmin、User
- 不引入 dotenv；仅使用现有环境变量；不覆盖 DB_PASSWORD / JWT_SECRET
- API 统一 /api/...；不出现 /api/v1、/api/v2

## Functional

- Windows + PowerShell 下可执行超管初始化命令
- 初始化命令使用现有 DB_* 环境变量
- SuperAdmin 可完成 MVP 主链路（创建/加入/授权/撤销）端到端可用
- 首次登录强制改密生效
- User 无法访问管理页面与管理 API
- 普通用户只能看到“我的可访问页面与资源”页
- `GET /api/permissions/resources` 返回当前用户有效 page 列表
- page 只管理页面资源（不混入 host/service）
- host 与 service 独立页面管理
- 页面资源管理路由为 /admin/pages，/admin/resources 仅作兼容重定向
- 页面资源管理支持路由扫描预览与同步注册（新增/变更/下线）
- page 主来源为 router/menu，手工仅保留补录异常路由
- page 列表展示 source 与授权数，支持跳转权限管理过滤
- 资源组管理 type 仅 host/service（不允许 page）
- environment / location 支持完整 CRUD
- 权限判定：
  - 用户直授、用户组直授均生效
  - ReadWrite 覆盖 ReadOnly
  - 撤销授权后结果立即更新
- Grant 升级/降级通过更新完成，不产生重复有效授权
- service 授权前置 host ReadOnly 生效
- credential 权限前置父资源 ReadOnly 生效
- credential 列表仅在父资源 ReadOnly 前置满足时可访问
- credential 明文查看要求父资源 ReadOnly + credential ReadOnly 以上
- credential 新增/修改/删除要求父资源 ReadOnly + credential ReadWrite
- credential 对象 permission 枚举仍仅 ReadOnly/ReadWrite（不新增第三套权限值）
- 普通用户页面访问命中基于 permissions/resources.route_path
- service 环境继承规则生效
- 权限管理新增弹窗不出现 subject_id/object_id 手填输入框
- 权限管理支持按名称搜索并选择主体/客体
- 权限列表展示主体名称/客体名称，不以裸 ID 为主要信息
- 权限管理新增授权显示“主体-客体-权限”预览文案
- 权限管理支持编辑授权（最小能力修改 permission）
- 用户组新增/编辑支持成员多选并保存
- 用户组列表显示成员数并可查看成员名称
- 资源组新增/编辑支持按类型联动的成员多选并保存
- 资源组列表显示成员数并可查看成员名称
- host 位置唯一绑定生效
- host/service credential 列表不返回明文 secret/cert/key
- reveal 使用单独接口并写审计
- 主机环境标签绑定弹窗支持查看/添加/移除
- 服务环境标签绑定弹窗支持查看/添加/移除并显示继承来源
- 列表能力：
  - 分页：users/user-groups/pages/hosts/services/environments/locations/grants
  - 筛选：按主体/客体/权限集合（最小集）
  - 搜索：按名称关键字（最小集）

## Security

- 密码不明文存储
- 登录后发放 JWT；接口鉴权默认开启
- 不返回敏感字段（例如密码哈希、密钥等）
- 凭据密文存储（证书/私钥密文字段为 LONGTEXT）
- 关键管理操作写入审计日志
- 审计日志记录成功/失败与变更前后摘要
- 普通用户页面与 API 越权验证通过

## Windows-first

- 本地启动与联调步骤在 Windows + PowerShell 下可执行
- 路径/脚本不依赖 bash 特性
