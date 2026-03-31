package db

import (
	"database/sql"
	"fmt"
	"strings"
)

func Migrate(conn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(64) NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			role ENUM('super_admin','user') NOT NULL DEFAULT 'user',
			status ENUM('enabled','disabled') NOT NULL DEFAULT 'enabled',
			must_change_password TINYINT(1) NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			UNIQUE KEY uq_users_username (username),
			KEY idx_users_role (role),
			KEY idx_users_status (status)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS user_groups (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(128) NOT NULL,
			description VARCHAR(255) NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			UNIQUE KEY uq_user_groups_name (name)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS user_group_members (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT NOT NULL,
			user_group_id BIGINT NOT NULL,
			created_at DATETIME NOT NULL,
			UNIQUE KEY uq_ugm_user_group (user_id, user_group_id),
			KEY idx_ugm_group (user_group_id),
			CONSTRAINT fk_ugm_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			CONSTRAINT fk_ugm_group FOREIGN KEY (user_group_id) REFERENCES user_groups(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS resources (
			` + "`key`" + ` BIGINT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(128) NOT NULL,
			type ENUM('host','service','page','api','business') NOT NULL,
			address VARCHAR(255) NULL,
			service_identifier VARCHAR(255) NULL,
			route_path VARCHAR(255) NULL,
			status ENUM('active','inactive') NOT NULL DEFAULT 'active',
			description VARCHAR(255) NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			KEY idx_resources_type (type),
			KEY idx_resources_status (status),
			KEY idx_resources_name (name)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='compatibility storage: resources';`,
		`CREATE TABLE IF NOT EXISTS pages (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(128) NOT NULL,
			route_path VARCHAR(255) NOT NULL,
			source ENUM('router','menu','manual') NOT NULL DEFAULT 'manual',
			status ENUM('active','inactive') NOT NULL DEFAULT 'active',
			description VARCHAR(255) NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			UNIQUE KEY uq_pages_route_path (route_path),
			KEY idx_pages_status (status),
			KEY idx_pages_name (name)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS locations (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			code VARCHAR(64) NOT NULL,
			name VARCHAR(128) NOT NULL,
			location_type ENUM('room','rack','idc','region','office','other') NOT NULL,
			address VARCHAR(255) NULL,
			status ENUM('active','inactive') NOT NULL DEFAULT 'active',
			description VARCHAR(255) NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			UNIQUE KEY uq_locations_code (code),
			KEY idx_locations_status (status),
			KEY idx_locations_name (name)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS hosts (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(128) NOT NULL,
			hostname VARCHAR(128) NOT NULL,
			primary_address VARCHAR(255) NOT NULL,
			provider_kind ENUM('physical','vm','cloud_instance','other') NOT NULL,
			cloud_vendor VARCHAR(64) NULL,
			cloud_instance_id VARCHAR(128) NULL,
			os_type VARCHAR(64) NULL,
			status ENUM('active','inactive') NOT NULL DEFAULT 'active',
			location_id BIGINT NULL,
			description VARCHAR(255) NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			KEY idx_hosts_location_id (location_id),
			KEY idx_hosts_status (status),
			KEY idx_hosts_name (name),
			CONSTRAINT fk_hosts_location FOREIGN KEY (location_id) REFERENCES locations(id) ON DELETE SET NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS services (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(128) NOT NULL,
			service_kind ENUM('app','api','database','middleware','cloud_product','other') NOT NULL,
			host_id BIGINT NULL,
			endpoint_or_identifier VARCHAR(255) NOT NULL,
			port INT NULL,
			protocol VARCHAR(32) NULL,
			cloud_vendor VARCHAR(64) NULL,
			cloud_product_code VARCHAR(128) NULL,
			status ENUM('active','inactive') NOT NULL DEFAULT 'active',
			description VARCHAR(255) NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			KEY idx_services_host_id (host_id),
			KEY idx_services_status (status),
			KEY idx_services_name (name),
			CONSTRAINT fk_services_host FOREIGN KEY (host_id) REFERENCES hosts(id) ON DELETE SET NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS environments (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			code VARCHAR(64) NOT NULL,
			name VARCHAR(128) NOT NULL,
			status ENUM('active','inactive') NOT NULL DEFAULT 'active',
			description VARCHAR(255) NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			UNIQUE KEY uq_environments_code (code),
			KEY idx_environments_status (status),
			KEY idx_environments_name (name)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS host_environments (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			host_id BIGINT NOT NULL,
			environment_id BIGINT NOT NULL,
			created_at DATETIME NOT NULL,
			UNIQUE KEY uq_host_env (host_id, environment_id),
			KEY idx_host_env_environment_id (environment_id),
			CONSTRAINT fk_host_env_host FOREIGN KEY (host_id) REFERENCES hosts(id) ON DELETE CASCADE,
			CONSTRAINT fk_host_env_environment FOREIGN KEY (environment_id) REFERENCES environments(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS service_environments (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			service_id BIGINT NOT NULL,
			environment_id BIGINT NOT NULL,
			created_at DATETIME NOT NULL,
			UNIQUE KEY uq_service_env (service_id, environment_id),
			KEY idx_service_env_environment_id (environment_id),
			CONSTRAINT fk_service_env_service FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE,
			CONSTRAINT fk_service_env_environment FOREIGN KEY (environment_id) REFERENCES environments(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS host_credentials (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			host_id BIGINT NOT NULL,
			account_name VARCHAR(128) NOT NULL,
			credential_name VARCHAR(128) NOT NULL,
			credential_kind ENUM('password','certificate') NOT NULL,
			username VARCHAR(128) NULL,
			secret_ciphertext LONGTEXT NULL,
			certificate_pem_ciphertext LONGTEXT NULL,
			private_key_pem_ciphertext LONGTEXT NULL,
			passphrase_ciphertext LONGTEXT NULL,
			status ENUM('active','inactive') NOT NULL DEFAULT 'active',
			description VARCHAR(255) NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			KEY idx_host_credentials_host_id (host_id),
			KEY idx_host_credentials_name (credential_name),
			CONSTRAINT fk_host_credentials_host FOREIGN KEY (host_id) REFERENCES hosts(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS service_credentials (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			service_id BIGINT NOT NULL,
			account_name VARCHAR(128) NOT NULL,
			credential_name VARCHAR(128) NOT NULL,
			credential_kind ENUM('password','certificate') NOT NULL,
			username VARCHAR(128) NULL,
			secret_ciphertext LONGTEXT NULL,
			certificate_pem_ciphertext LONGTEXT NULL,
			private_key_pem_ciphertext LONGTEXT NULL,
			passphrase_ciphertext LONGTEXT NULL,
			status ENUM('active','inactive') NOT NULL DEFAULT 'active',
			description VARCHAR(255) NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			KEY idx_service_credentials_service_id (service_id),
			KEY idx_service_credentials_name (credential_name),
			CONSTRAINT fk_service_credentials_service FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS resource_groups (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(128) NOT NULL,
			type ENUM('host','service') NOT NULL,
			description VARCHAR(255) NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			UNIQUE KEY uq_resource_groups_name (name),
			KEY idx_resource_groups_type (type)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='compatibility storage: resource_groups';`,
		`CREATE TABLE IF NOT EXISTS resource_group_members (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			resource_key BIGINT NOT NULL,
			resource_group_id BIGINT NOT NULL,
			created_at DATETIME NOT NULL,
			UNIQUE KEY uq_rgm_resource_group (resource_key, resource_group_id),
			KEY idx_rgm_group (resource_group_id),
			CONSTRAINT fk_rgm_resource FOREIGN KEY (resource_key) REFERENCES resources(` + "`key`" + `) ON DELETE CASCADE,
			CONSTRAINT fk_rgm_group FOREIGN KEY (resource_group_id) REFERENCES resource_groups(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='compatibility storage: resource_group_members';`,
		`CREATE TABLE IF NOT EXISTS grants (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			subject_type ENUM('user','user_group') NOT NULL,
			subject_id BIGINT NOT NULL,
			object_type ENUM('page','host','host_group','service','service_group','host_credential','service_credential') NOT NULL,
			object_id BIGINT NOT NULL,
			permission_code VARCHAR(64) NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			UNIQUE KEY uq_grants_subject_object_permission (subject_type, subject_id, object_type, object_id, permission_code),
			KEY idx_grants_subject (subject_type, subject_id),
			KEY idx_grants_object (object_type, object_id),
			KEY idx_grants_permission_code (permission_code)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS permission_definitions (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			code VARCHAR(64) NOT NULL,
			object_family ENUM('page','host','service','host_credential','service_credential') NOT NULL,
			action VARCHAR(32) NOT NULL,
			display_name VARCHAR(128) NOT NULL,
			description VARCHAR(255) NULL,
			status ENUM('active','inactive') NOT NULL DEFAULT 'active',
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			UNIQUE KEY uq_permission_definitions_code (code),
			KEY idx_permission_definitions_family (object_family),
			KEY idx_permission_definitions_status (status)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			actor_user_id BIGINT NOT NULL,
			actor_username_snapshot VARCHAR(64) NOT NULL,
			action VARCHAR(64) NOT NULL,
			target_type VARCHAR(64) NOT NULL,
			target_id VARCHAR(64) NOT NULL,
			target_name_snapshot VARCHAR(128) NOT NULL,
			occurred_at DATETIME NOT NULL,
			before_json JSON NULL,
			after_json JSON NULL,
			result ENUM('success','failure') NOT NULL,
			ip VARCHAR(64) NULL,
			KEY idx_audit_actor (actor_user_id),
			KEY idx_audit_target (target_type, target_id),
			KEY idx_audit_occurred_at (occurred_at),
			KEY idx_audit_result (result),
			CONSTRAINT fk_audit_actor FOREIGN KEY (actor_user_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
	}
	for _, stmt := range stmts {
		if _, err := conn.Exec(stmt); err != nil {
			return err
		}
	}
	if err := ensureColumn(conn, "user_groups", "description", `ALTER TABLE user_groups ADD COLUMN description VARCHAR(255) NULL`); err != nil {
		return err
	}
	if err := ensureColumn(conn, "resource_groups", "type", `ALTER TABLE resource_groups ADD COLUMN type ENUM('host','service') NOT NULL DEFAULT 'host'`); err != nil {
		return err
	}
	if err := ensureColumn(conn, "pages", "source", `ALTER TABLE pages ADD COLUMN source ENUM('router','menu','manual') NOT NULL DEFAULT 'manual' AFTER route_path`); err != nil {
		return err
	}
	if _, err := conn.Exec(`ALTER TABLE resources MODIFY COLUMN type ENUM('host','service','page','api','business') NOT NULL`); err != nil && !ignorableMigrationErr(err) {
		return err
	}
	if _, err := conn.Exec(`ALTER TABLE grants MODIFY COLUMN object_type ENUM('resource','resource_group','page','host','host_group','service','service_group','host_credential','service_credential') NOT NULL`); err != nil && !ignorableMigrationErr(err) {
		return err
	}
	if err := ensureColumn(conn, "grants", "permission_code", `ALTER TABLE grants ADD COLUMN permission_code VARCHAR(64) NULL AFTER object_id`); err != nil {
		return err
	}
	hasOldPermission, err := columnExists(conn, "grants", "permission")
	if err != nil {
		return err
	}
	if hasOldPermission {
		if _, err := conn.Exec(`UPDATE grants SET permission_code=permission WHERE (permission_code IS NULL OR permission_code='') AND permission IS NOT NULL`); err != nil && !ignorableMigrationErr(err) {
			return err
		}
	}
	if _, err := conn.Exec(`
UPDATE grants g
LEFT JOIN resource_groups rg ON rg.id=g.object_id
SET g.object_type=CASE WHEN rg.type='service' THEN 'service_group' ELSE 'host_group' END
WHERE g.object_type='resource_group'`); err != nil && !ignorableMigrationErr(err) {
		return err
	}
	if _, err := conn.Exec(`
UPDATE grants g
LEFT JOIN resources r ON r.` + "`key`" + `=g.object_id
SET g.object_type=CASE WHEN r.type='service' THEN 'service' ELSE 'host' END
WHERE g.object_type='resource'`); err != nil && !ignorableMigrationErr(err) {
		return err
	}
	if _, err := conn.Exec(`
UPDATE grants
SET permission_code = CASE
	WHEN object_type='page' THEN 'page.view'
	WHEN object_type IN ('host','host_group') AND permission_code='ReadOnly' THEN 'host.read'
	WHEN object_type IN ('host','host_group') AND permission_code='ReadWrite' THEN 'host.write'
	WHEN object_type IN ('service','service_group') AND permission_code='ReadOnly' THEN 'service.read'
	WHEN object_type IN ('service','service_group') AND permission_code='ReadWrite' THEN 'service.write'
	WHEN object_type='host_credential' AND permission_code='ReadOnly' THEN 'host_credential.read'
	WHEN object_type='host_credential' AND permission_code='ReadWrite' THEN 'host_credential.write'
	WHEN object_type='service_credential' AND permission_code='ReadOnly' THEN 'service_credential.read'
	WHEN object_type='service_credential' AND permission_code='ReadWrite' THEN 'service_credential.write'
	ELSE permission_code
END`); err != nil && !ignorableMigrationErr(err) {
		return err
	}
	if _, err := conn.Exec(`DELETE FROM grants WHERE object_type NOT IN ('page','host','host_group','service','service_group','host_credential','service_credential')`); err != nil && !ignorableMigrationErr(err) {
		return err
	}
	if _, err := conn.Exec(`ALTER TABLE grants MODIFY COLUMN object_type ENUM('page','host','host_group','service','service_group','host_credential','service_credential') NOT NULL`); err != nil && !ignorableMigrationErr(err) {
		return err
	}
	if _, err := conn.Exec(`ALTER TABLE grants MODIFY COLUMN permission_code VARCHAR(64) NOT NULL`); err != nil && !ignorableMigrationErr(err) {
		return err
	}
	if hasOldPermission {
		if _, err := conn.Exec(`ALTER TABLE grants DROP COLUMN permission`); err != nil && !ignorableMigrationErr(err) {
			return err
		}
	}
	if _, err := conn.Exec(`CREATE INDEX idx_grants_permission_code ON grants(permission_code)`); err != nil && !ignorableMigrationErr(err) {
		return err
	}
	oldGrantUnique, err := indexExists(conn, "grants", "uq_grants_subject_object")
	if err != nil {
		return err
	}
	newGrantUnique, err := indexExists(conn, "grants", "uq_grants_subject_object_permission")
	if err != nil {
		return err
	}
	if oldGrantUnique {
		if _, err := conn.Exec(`ALTER TABLE grants DROP INDEX uq_grants_subject_object`); err != nil && !ignorableMigrationErr(err) {
			return err
		}
	}
	if !newGrantUnique {
		if _, err := conn.Exec(`ALTER TABLE grants ADD UNIQUE INDEX uq_grants_subject_object_permission (subject_type, subject_id, object_type, object_id, permission_code)`); err != nil && !ignorableMigrationErr(err) {
			return err
		}
	}
	if _, err := conn.Exec(`
INSERT INTO permission_definitions(code,object_family,action,display_name,description,status,created_at,updated_at) VALUES
('page.view','page','view','页面访问','页面可见与路由可访问','active',NOW(),NOW()),
('host.read','host','read','主机只读','查看主机及普通元数据','active',NOW(),NOW()),
('host.write','host','write','主机读写','主机写操作','active',NOW(),NOW()),
('service.read','service','read','服务只读','查看服务及普通元数据','active',NOW(),NOW()),
('service.write','service','write','服务读写','服务写操作','active',NOW(),NOW()),
('host_credential.read','host_credential','read','主机凭据只读','查看主机凭据元数据','active',NOW(),NOW()),
('host_credential.write','host_credential','write','主机凭据读写','主机凭据写操作','active',NOW(),NOW()),
('service_credential.read','service_credential','read','服务凭据只读','查看服务凭据元数据','active',NOW(),NOW()),
('service_credential.write','service_credential','write','服务凭据读写','服务凭据写操作','active',NOW(),NOW())
ON DUPLICATE KEY UPDATE
object_family=VALUES(object_family),
action=VALUES(action),
display_name=VALUES(display_name),
description=VALUES(description),
status=VALUES(status),
updated_at=NOW()`); err != nil && !ignorableMigrationErr(err) {
		return err
	}
	if _, err := conn.Exec(`
INSERT INTO grants(subject_type,subject_id,object_type,object_id,permission_code,created_at,updated_at)
SELECT
	g.subject_type,
	g.subject_id,
	g.object_type,
	g.object_id,
	CASE
		WHEN g.permission_code='host_credential.reveal' THEN 'host_credential.read'
		WHEN g.permission_code='service_credential.reveal' THEN 'service_credential.read'
		ELSE g.permission_code
	END AS permission_code,
	NOW(),
	NOW()
FROM grants g
WHERE g.permission_code IN ('host_credential.reveal','service_credential.reveal')
AND NOT EXISTS (
	SELECT 1
	FROM grants g2
	WHERE g2.subject_type=g.subject_type
		AND g2.subject_id=g.subject_id
		AND g2.object_type=g.object_type
		AND g2.object_id=g.object_id
		AND g2.permission_code=CASE
			WHEN g.permission_code='host_credential.reveal' THEN 'host_credential.read'
			WHEN g.permission_code='service_credential.reveal' THEN 'service_credential.read'
			ELSE g.permission_code
		END
)`); err != nil && !ignorableMigrationErr(err) {
		return err
	}
	if _, err := conn.Exec(`DELETE FROM grants WHERE permission_code IN ('host_credential.reveal','service_credential.reveal')`); err != nil && !ignorableMigrationErr(err) {
		return err
	}
	if _, err := conn.Exec(`DELETE FROM permission_definitions WHERE code IN ('host_credential.reveal','service_credential.reveal')`); err != nil && !ignorableMigrationErr(err) {
		return err
	}
	if _, err := conn.Exec(`ALTER TABLE resource_groups MODIFY COLUMN type ENUM('host','service') NOT NULL`); err != nil && !ignorableMigrationErr(err) {
		return err
	}
	if err := ensureColumn(conn, "resources", "address", `ALTER TABLE resources ADD COLUMN address VARCHAR(255) NULL`); err != nil {
		return err
	}
	if err := ensureColumn(conn, "resources", "service_identifier", `ALTER TABLE resources ADD COLUMN service_identifier VARCHAR(255) NULL`); err != nil {
		return err
	}
	if err := ensureColumn(conn, "resources", "route_path", `ALTER TABLE resources ADD COLUMN route_path VARCHAR(255) NULL`); err != nil {
		return err
	}
	if err := ensureColumn(conn, "resource_groups", "description", `ALTER TABLE resource_groups ADD COLUMN description VARCHAR(255) NULL`); err != nil {
		return err
	}
	if err := ensureColumn(conn, "resource_group_members", "resource_key", `ALTER TABLE resource_group_members ADD COLUMN resource_key BIGINT NULL`); err != nil {
		return err
	}
	hasOld, err := columnExists(conn, "resource_group_members", "resource_id")
	if err != nil {
		return err
	}
	if hasOld {
		if _, err := conn.Exec(`UPDATE resource_group_members SET resource_key = resource_id WHERE resource_key IS NULL AND resource_id IS NOT NULL`); err != nil && !ignorableMigrationErr(err) {
			return err
		}
	}
	if _, err := conn.Exec(`ALTER TABLE resource_group_members MODIFY COLUMN resource_key BIGINT NOT NULL`); err != nil && !ignorableMigrationErr(err) {
		return err
	}
	if _, err := conn.Exec(`CREATE UNIQUE INDEX uq_rgm_resource_group ON resource_group_members(resource_key, resource_group_id)`); err != nil && !ignorableMigrationErr(err) {
		return err
	}
	if err := normalizeCharset(conn); err != nil {
		return err
	}
	return nil
}

func normalizeCharset(conn *sql.DB) error {
	var dbName sql.NullString
	if err := conn.QueryRow(`SELECT DATABASE()`).Scan(&dbName); err != nil {
		return err
	}
	if dbName.Valid && dbName.String != "" {
		stmt := fmt.Sprintf("ALTER DATABASE `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci", strings.ReplaceAll(dbName.String, "`", "``"))
		if _, err := conn.Exec(stmt); err != nil && !ignorableMigrationErr(err) {
			return err
		}
	}
	tables := []string{
		"users", "user_groups", "user_group_members", "resources", "pages", "locations",
		"hosts", "services", "environments", "host_environments", "service_environments",
		"host_credentials", "service_credentials", "resource_groups", "resource_group_members",
		"grants", "permission_definitions", "audit_logs",
	}
	for _, table := range tables {
		stmt := fmt.Sprintf("ALTER TABLE `%s` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci", table)
		if _, err := conn.Exec(stmt); err != nil && !ignorableMigrationErr(err) {
			return err
		}
	}
	return nil
}

func ensureColumn(conn *sql.DB, tableName, columnName, alterSQL string) error {
	ok, err := columnExists(conn, tableName, columnName)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	_, err = conn.Exec(alterSQL)
	return err
}

func columnExists(conn *sql.DB, tableName, columnName string) (bool, error) {
	var cnt int
	err := conn.QueryRow(`SELECT COUNT(1) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ? AND column_name = ?`, tableName, columnName).Scan(&cnt)
	return cnt > 0, err
}

func indexExists(conn *sql.DB, tableName, indexName string) (bool, error) {
	var cnt int
	err := conn.QueryRow(`SELECT COUNT(1) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = ? AND index_name = ?`, tableName, indexName).Scan(&cnt)
	return cnt > 0, err
}

func ignorableMigrationErr(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key name") ||
		strings.Contains(msg, "duplicate column name") ||
		strings.Contains(msg, "already exists")
}
