package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type textColumn struct {
	table  string
	column string
}

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
	if err := printDBCharset(db); err != nil {
		panic(err)
	}
	if err := printConnectionCharsets(db); err != nil {
		panic(err)
	}
	if err := printTableCollations(db); err != nil {
		panic(err)
	}
	if err := printColumnCharsets(db); err != nil {
		panic(err)
	}
	if err := printQuestionMarkDataHints(db); err != nil {
		panic(err)
	}
}

func printConnectionCharsets(db *sql.DB) error {
	rows, err := db.Query(`SHOW VARIABLES WHERE Variable_name IN (
'character_set_client','character_set_connection','character_set_database',
'character_set_results','character_set_server','collation_connection','collation_database','collation_server'
)`)
	if err != nil {
		return err
	}
	defer rows.Close()
	fmt.Println("CONNECTION charset variables:")
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return err
		}
		fmt.Printf("- %s=%s\n", k, v)
	}
	return rows.Err()
}

func printDBCharset(db *sql.DB) error {
	var cs, coll string
	err := db.QueryRow(`
SELECT DEFAULT_CHARACTER_SET_NAME, DEFAULT_COLLATION_NAME
FROM information_schema.SCHEMATA
WHERE SCHEMA_NAME = DATABASE()`).Scan(&cs, &coll)
	if err != nil {
		return err
	}
	fmt.Printf("DATABASE charset=%s collation=%s\n", cs, coll)
	return nil
}

func printTableCollations(db *sql.DB) error {
	rows, err := db.Query(`
SELECT TABLE_NAME, TABLE_COLLATION
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = DATABASE()
ORDER BY TABLE_NAME`)
	if err != nil {
		return err
	}
	defer rows.Close()
	var bad []string
	for rows.Next() {
		var t, c string
		if err := rows.Scan(&t, &c); err != nil {
			return err
		}
		if !strings.HasPrefix(strings.ToLower(c), "utf8mb4_") {
			bad = append(bad, fmt.Sprintf("%s(%s)", t, c))
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(bad) == 0 {
		fmt.Println("TABLE collation check=PASS(all utf8mb4_*)")
		return nil
	}
	fmt.Printf("TABLE collation check=FAIL non-utf8mb4 tables: %s\n", strings.Join(bad, ", "))
	return nil
}

func printColumnCharsets(db *sql.DB) error {
	rows, err := db.Query(`
SELECT TABLE_NAME, COLUMN_NAME, CHARACTER_SET_NAME, COLLATION_NAME
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
  AND DATA_TYPE IN ('char','varchar','text','tinytext','mediumtext','longtext')
ORDER BY TABLE_NAME, ORDINAL_POSITION`)
	if err != nil {
		return err
	}
	defer rows.Close()
	var bad []string
	for rows.Next() {
		var t, c, cs, coll sql.NullString
		if err := rows.Scan(&t, &c, &cs, &coll); err != nil {
			return err
		}
		if !cs.Valid || !strings.EqualFold(cs.String, "utf8mb4") {
			bad = append(bad, fmt.Sprintf("%s.%s(cs=%s,coll=%s)", t.String, c.String, cs.String, coll.String))
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(bad) == 0 {
		fmt.Println("COLUMN charset check=PASS(all text columns utf8mb4)")
		return nil
	}
	fmt.Printf("COLUMN charset check=FAIL non-utf8mb4 columns: %s\n", strings.Join(bad, ", "))
	return nil
}

func printQuestionMarkDataHints(db *sql.DB) error {
	targets := []textColumn{
		{"users", "username"},
		{"user_groups", "name"},
		{"pages", "name"},
		{"resources", "name"},
		{"resource_groups", "name"},
		{"hosts", "name"},
		{"services", "name"},
		{"environments", "name"},
		{"locations", "name"},
		{"host_credentials", "account_name"},
		{"service_credentials", "account_name"},
	}
	fmt.Println("QUESTION_MARK data scan(hint only):")
	for _, t := range targets {
		q := fmt.Sprintf("SELECT COUNT(1) FROM `%s` WHERE `%s` LIKE '%%?%%'", t.table, t.column)
		var cnt int
		if err := db.QueryRow(q).Scan(&cnt); err != nil {
			return err
		}
		if cnt == 0 {
			continue
		}
		var id int64
		var val, hexVal string
		sample := fmt.Sprintf("SELECT id, `%s`, HEX(`%s`) FROM `%s` WHERE `%s` LIKE '%%?%%' ORDER BY id DESC LIMIT 1", t.column, t.column, t.table, t.column)
		if err := db.QueryRow(sample).Scan(&id, &val, &hexVal); err != nil {
			return err
		}
		fmt.Printf("- %s.%s count=%d sample_id=%d sample_value=%s sample_hex=%s\n", t.table, t.column, cnt, id, val, hexVal)
	}
	return nil
}
