package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	name := os.Getenv("DB_NAME")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	if host == "" || port == "" || name == "" || user == "" || pass == "" {
		panic("missing DB_HOST/DB_PORT/DB_NAME/DB_USER/DB_PASSWORD")
	}
	if !strings.EqualFold(name, "irms") {
		panic("DB_NAME must be irms")
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=true&loc=Local", user, pass, host, port, name)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		panic(err)
	}
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()
	if err := cleanupBadData(tx); err != nil {
		panic(err)
	}
	if err := insertFreshSamples(tx); err != nil {
		panic(err)
	}
	if err := tx.Commit(); err != nil {
		panic(err)
	}
	fmt.Println("reset done")
}

func cleanupBadData(tx *sql.Tx) error {
	pageIDs, err := collectIDs(tx, "SELECT id FROM pages WHERE name LIKE '%?%'")
	if err != nil {
		return err
	}
	envIDs, err := collectIDs(tx, "SELECT id FROM environments WHERE name LIKE '%?%'")
	if err != nil {
		return err
	}
	locIDs, err := collectIDs(tx, "SELECT id FROM locations WHERE name LIKE '%?%'")
	if err != nil {
		return err
	}
	for _, id := range pageIDs {
		if _, err := tx.Exec("DELETE FROM grants WHERE object_type='page' AND object_id=?", id); err != nil {
			return err
		}
		if _, err := tx.Exec("DELETE FROM pages WHERE id=?", id); err != nil {
			return err
		}
	}
	for _, id := range envIDs {
		if _, err := tx.Exec("DELETE FROM host_environments WHERE environment_id=?", id); err != nil {
			return err
		}
		if _, err := tx.Exec("DELETE FROM service_environments WHERE environment_id=?", id); err != nil {
			return err
		}
		if _, err := tx.Exec("DELETE FROM environments WHERE id=?", id); err != nil {
			return err
		}
	}
	for _, id := range locIDs {
		if _, err := tx.Exec("UPDATE hosts SET location_id=NULL WHERE location_id=?", id); err != nil {
			return err
		}
		if _, err := tx.Exec("DELETE FROM locations WHERE id=?", id); err != nil {
			return err
		}
	}
	fmt.Printf("deleted pages=%d environments=%d locations=%d\n", len(pageIDs), len(envIDs), len(locIDs))
	return nil
}

func insertFreshSamples(tx *sql.Tx) error {
	now := time.Now()
	sfx := now.Format("150405")
	if _, err := tx.Exec(
		"INSERT INTO pages(name,route_path,status,description,created_at,updated_at) VALUES(?,?,?,?,?,?)",
		zh(20013, 25991, 39029, 38754)+"-"+sfx, "/zh-page-"+sfx, "active", zh(24320, 21457, 29615, 22659, 20013, 25991, 26679, 20363), now, now,
	); err != nil {
		return err
	}
	envRes, err := tx.Exec(
		"INSERT INTO environments(code,name,status,description,created_at,updated_at) VALUES(?,?,?,?,?,?)",
		"zh-env-"+sfx, zh(29983, 20135, 29615, 22659)+"-"+sfx, "active", zh(24320, 21457, 29615, 22659, 20013, 25991, 26679, 20363), now, now,
	)
	if err != nil {
		return err
	}
	envID, err := envRes.LastInsertId()
	if err != nil {
		return err
	}
	locRes, err := tx.Exec(
		"INSERT INTO locations(code,name,location_type,address,status,description,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?)",
		"zh-loc-"+sfx, zh(21271, 20140, 26426, 25151)+"-"+sfx, "idc", zh(21271, 20140, 24066, 26397, 38451, 38451, 21306), "active", zh(24320, 21457, 29615, 22659, 20013, 25991, 26679, 20363), now, now,
	)
	if err != nil {
		return err
	}
	locID, err := locRes.LastInsertId()
	if err != nil {
		return err
	}
	hostRes, err := tx.Exec(
		"INSERT INTO hosts(name,hostname,primary_address,provider_kind,cloud_vendor,cloud_instance_id,os_type,status,location_id,description,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)",
		zh(20013, 25991, 20027, 26426)+"-"+sfx, "zh-host-"+sfx+".local", "10.20.30.40", "vm", nil, nil, "linux", "active", locID, zh(24320, 21457, 29615, 22659, 20013, 25991, 26679, 20363), now, now,
	)
	if err != nil {
		return err
	}
	hostID, err := hostRes.LastInsertId()
	if err != nil {
		return err
	}
	if _, err := tx.Exec("INSERT INTO host_environments(host_id,environment_id,created_at) VALUES(?,?,?)", hostID, envID, now); err != nil {
		return err
	}
	if _, err := tx.Exec(
		"INSERT INTO services(name,service_kind,host_id,endpoint_or_identifier,port,protocol,cloud_vendor,cloud_product_code,status,description,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)",
		zh(20013, 25991, 26381, 21153)+"-"+sfx, "api", hostID, "zh-service-"+sfx, 8081, "http", nil, nil, "active", zh(24320, 21457, 29615, 22659, 20013, 25991, 26679, 20363), now, now,
	); err != nil {
		return err
	}
	fmt.Println("inserted fresh Chinese samples")
	return nil
}

func zh(r ...rune) string {
	return string(r)
}

func collectIDs(tx *sql.Tx, query string) ([]int64, error) {
	rows, err := tx.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
