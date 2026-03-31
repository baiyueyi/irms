# IRMS RBAC Data Model (MVP)

## 1) 数据库与约束范围

- 数据库：MySQL
- 环境变量：仅使用 `DB_HOST` `DB_PORT` `DB_NAME` `DB_USER` `DB_PASSWORD`
- 不引入 dotenv，不覆盖已有环境变量
- 本文覆盖：RBAC + page/host/service + environments/locations + credentials 最小闭环

## 2) 枚举定义

- `user_role`: `super_admin` | `user`
- `user_status`: `enabled` | `disabled`
- `common_status`: `active` | `inactive`
- `subject_type`: `user` | `user_group`
- `object_type`: `page` | `host` | `service` | `host_group` | `service_group` | `host_credential` | `service_credential`
- `permission_code`: `page.view` | `host.read` | `host.write` | `service.read` | `service.write` | `host_credential.read` | `host_credential.write` | `service_credential.read` | `service_credential.write`
- `provider_kind`: `physical` | `vm` | `cloud_instance` | `other`
- `service_kind`: `app` | `api` | `database` | `middleware` | `cloud_product` | `other`
- `location_type`: `room` | `rack` | `idc` | `region` | `office` | `other`
- `credential_kind`: `password` | `certificate`
- `audit_result`: `success` | `failure`

## 3) 表清单

- `users`
- `user_groups`
- `user_group_members`
- `pages`
- `hosts`
- `services`
- `environments`
- `locations`
- `resource_groups`
- `resource_group_members`
- `host_environments`
- `service_environments`
- `host_credentials`
- `service_credentials`
- `grants`
- `audit_logs`

## 4) 表结构（最小实现）

### 4.1 pages

- `id` BIGINT PK AUTO_INCREMENT
- `name` VARCHAR(128) NOT NULL
- `route_path` VARCHAR(255) NOT NULL
- `source` ENUM('router','menu','manual') NOT NULL DEFAULT 'manual'
- `status` ENUM('active','inactive') NOT NULL DEFAULT 'active'
- `description` VARCHAR(255) NULL
- `created_at` DATETIME NOT NULL
- `updated_at` DATETIME NOT NULL
- `UNIQUE KEY uq_pages_route_path (route_path)`

### 4.2 locations

- `id` BIGINT PK AUTO_INCREMENT
- `code` VARCHAR(64) NOT NULL
- `name` VARCHAR(128) NOT NULL
- `location_type` ENUM('room','rack','idc','region','office','other') NOT NULL
- `address` VARCHAR(255) NULL
- `status` ENUM('active','inactive') NOT NULL DEFAULT 'active'
- `description` VARCHAR(255) NULL
- `created_at` DATETIME NOT NULL
- `updated_at` DATETIME NOT NULL
- `UNIQUE KEY uq_locations_code (code)`

### 4.3 hosts

- `id` BIGINT PK AUTO_INCREMENT
- `name` VARCHAR(128) NOT NULL
- `hostname` VARCHAR(128) NOT NULL
- `primary_address` VARCHAR(255) NOT NULL
- `provider_kind` ENUM('physical','vm','cloud_instance','other') NOT NULL
- `cloud_vendor` VARCHAR(64) NULL
- `cloud_instance_id` VARCHAR(128) NULL
- `os_type` VARCHAR(64) NULL
- `status` ENUM('active','inactive') NOT NULL DEFAULT 'active'
- `location_id` BIGINT NULL
- `description` VARCHAR(255) NULL
- `created_at` DATETIME NOT NULL
- `updated_at` DATETIME NOT NULL
- `FOREIGN KEY (location_id) REFERENCES locations(id)`

### 4.4 services

- `id` BIGINT PK AUTO_INCREMENT
- `name` VARCHAR(128) NOT NULL
- `service_kind` ENUM('app','api','database','middleware','cloud_product','other') NOT NULL
- `host_id` BIGINT NULL
- `endpoint_or_identifier` VARCHAR(255) NOT NULL
- `port` INT NULL
- `protocol` VARCHAR(32) NULL
- `cloud_vendor` VARCHAR(64) NULL
- `cloud_product_code` VARCHAR(128) NULL
- `status` ENUM('active','inactive') NOT NULL DEFAULT 'active'
- `description` VARCHAR(255) NULL
- `created_at` DATETIME NOT NULL
- `updated_at` DATETIME NOT NULL
- `FOREIGN KEY (host_id) REFERENCES hosts(id)`

### 4.5 environments

- `id` BIGINT PK AUTO_INCREMENT
- `code` VARCHAR(64) NOT NULL
- `name` VARCHAR(128) NOT NULL
- `status` ENUM('active','inactive') NOT NULL DEFAULT 'active'
- `description` VARCHAR(255) NULL
- `created_at` DATETIME NOT NULL
- `updated_at` DATETIME NOT NULL
- `UNIQUE KEY uq_environments_code (code)`

### 4.6 host_environments

- `id` BIGINT PK AUTO_INCREMENT
- `host_id` BIGINT NOT NULL
- `environment_id` BIGINT NOT NULL
- `created_at` DATETIME NOT NULL
- `UNIQUE KEY uq_host_env (host_id, environment_id)`

### 4.7 service_environments

- `id` BIGINT PK AUTO_INCREMENT
- `service_id` BIGINT NOT NULL
- `environment_id` BIGINT NOT NULL
- `created_at` DATETIME NOT NULL
- `UNIQUE KEY uq_service_env (service_id, environment_id)`

### 4.8 resource_groups（compatibility storage）

- `id` BIGINT PK AUTO_INCREMENT
- `name` VARCHAR(128) NOT NULL
- `type` ENUM('host','service') NOT NULL
- `description` VARCHAR(255) NULL
- `created_at` DATETIME NOT NULL
- `updated_at` DATETIME NOT NULL
- `UNIQUE KEY uq_resource_groups_name (name)`

说明：
- 资源组仅允许 host/service 两类
- page 不允许进入资源组
- 该表为 compatibility storage；正式对外语义为 `host_group/service_group`

### 4.9 resource_group_members（compatibility storage）

- `id` BIGINT PK AUTO_INCREMENT
- `resource_key` BIGINT NOT NULL
- `resource_group_id` BIGINT NOT NULL
- `created_at` DATETIME NOT NULL
- `UNIQUE KEY uq_rgm_resource_group (resource_key, resource_group_id)`

说明：
- 管理端新增/编辑资源组时可增删成员
- 成员展示使用资源名称（`resources.name`）
- 该表不作为正式 RBAC 主概念，仅承载 group 成员关系的存储实现

### 4.10 host_credentials

- `id` BIGINT PK AUTO_INCREMENT
- `host_id` BIGINT NOT NULL
- `account_name` VARCHAR(128) NOT NULL
- `credential_name` VARCHAR(128) NOT NULL
- `credential_kind` ENUM('password','certificate') NOT NULL
- `username` VARCHAR(128) NULL
- `secret_ciphertext` LONGTEXT NULL
- `certificate_pem_ciphertext` LONGTEXT NULL
- `private_key_pem_ciphertext` LONGTEXT NULL
- `passphrase_ciphertext` LONGTEXT NULL
- `status` ENUM('active','inactive') NOT NULL DEFAULT 'active'
- `description` VARCHAR(255) NULL
- `created_at` DATETIME NOT NULL
- `updated_at` DATETIME NOT NULL

### 4.11 service_credentials

- `id` BIGINT PK AUTO_INCREMENT
- `service_id` BIGINT NOT NULL
- `account_name` VARCHAR(128) NOT NULL
- `credential_name` VARCHAR(128) NOT NULL
- `credential_kind` ENUM('password','certificate') NOT NULL
- `username` VARCHAR(128) NULL
- `secret_ciphertext` LONGTEXT NULL
- `certificate_pem_ciphertext` LONGTEXT NULL
- `private_key_pem_ciphertext` LONGTEXT NULL
- `passphrase_ciphertext` LONGTEXT NULL
- `status` ENUM('active','inactive') NOT NULL DEFAULT 'active'
- `description` VARCHAR(255) NULL
- `created_at` DATETIME NOT NULL
- `updated_at` DATETIME NOT NULL

### 4.12 grants

- `id` BIGINT PK AUTO_INCREMENT
- `subject_type` ENUM('user','user_group') NOT NULL
- `subject_id` BIGINT NOT NULL
- `object_type` ENUM('page','host','service','host_group','service_group','host_credential','service_credential') NOT NULL
- `object_id` BIGINT NOT NULL
- `permission_code` VARCHAR(64) NOT NULL
- `created_at` DATETIME NOT NULL
- `updated_at` DATETIME NOT NULL
- `UNIQUE KEY uq_grants_subject_object_permission (subject_type, subject_id, object_type, object_id, permission_code)`

说明：
- 管理端授权 UI 语义客体类型：`page | host | service | host_group | service_group`
- 管理端交互不允许手填 `subject_id/object_id`，必须按名称搜索后选择
- 授权编辑最小能力：按 grant id 更新 `permission_code`
- 凭据权限码仅保留 `host_credential.read/write`、`service_credential.read/write`，不再使用 `*_credential.reveal`

### 4.13 audit_logs

- `id` BIGINT PK AUTO_INCREMENT
- `actor_user_id` BIGINT NOT NULL
- `actor_username_snapshot` VARCHAR(64) NOT NULL
- `action` VARCHAR(64) NOT NULL
- `target_type` VARCHAR(64) NOT NULL
- `target_id` VARCHAR(64) NOT NULL
- `target_name_snapshot` VARCHAR(128) NOT NULL
- `occurred_at` DATETIME NOT NULL
- `before_json` JSON NULL
- `after_json` JSON NULL
- `result` ENUM('success','failure') NOT NULL
- `ip` VARCHAR(64) NULL

## 5) 关键关系与规则

- `page` 仅用于页面资源授权与页面访问控制
- page 主来源是前端 router/menu 同步注册；手工补录仅用于异常路由
- host 与 service 分离：一个 host 可有多个 services；`services.host_id` 可为空
- host 与 location：一个 host 仅一个 location；service/page 不绑定 location
- host 与 environment：多对多
- service 与 environment：多对多
- `resources/resource_groups/resource_group_members` 为 compatibility storage，不作为正式授权主链路概念
- service 环境标签生效：
  - service 有标签，用 service 标签
  - service 无标签且 host_id 非空，继承 host 标签
  - service 无标签且 host_id 为空，无标签

## 6) 凭据与安全要求

- 凭据存独立表：`host_credentials` / `service_credentials`
- 不在 host/service 主表存凭据明文
- 明文密码/证书/私钥不落库
- 证书与私钥密文使用 `LONGTEXT`
- 列表接口不返回明文字段
- 明文查看必须经单独接口并记录审计日志

## 7) 权限前置规则

- service 授权前置：当 `services.host_id` 非空，授予 `service.read/service.write` 前需已有 host `host.read` 以上权限
- credential 权限前置：
  - host_credential 需先有对应 host `host.read` 以上
  - service_credential 需先有对应 service `service.read` 以上
- credential 动作级最小策略：
  - 列表：父资源 `host.read/service.read` 即可
  - 明文查看：父资源 read + 对应 credential 对象 `*.read`（`*.write` 隐含）
  - 新增/修改/删除：父资源 read + 对应 credential 对象 `*.write`
- credential 对象权限码：
  - `host_credential.read` / `host_credential.write`
  - `service_credential.read` / `service_credential.write`
