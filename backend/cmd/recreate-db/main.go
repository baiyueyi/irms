package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"irms/backend/internal/db"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	if host == "" || port == "" || user == "" || pass == "" {
		panic("missing DB_HOST/DB_PORT/DB_USER/DB_PASSWORD")
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		panic("missing DB_NAME")
	}
	if !strings.EqualFold(dbName, "irms") {
		panic("DB_NAME must be irms")
	}
	baseDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=true&loc=Local&multiStatements=true", user, pass, host, port)
	baseConn, err := sql.Open("mysql", baseDSN)
	if err != nil {
		panic(err)
	}
	defer baseConn.Close()

	if _, err := baseConn.Exec("DROP DATABASE IF EXISTS `" + dbName + "`; CREATE DATABASE `" + dbName + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;"); err != nil {
		panic(err)
	}

	targetDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=true&loc=Local", user, pass, host, port, dbName)
	targetConn, err := sql.Open("mysql", targetDSN)
	if err != nil {
		panic(err)
	}
	defer targetConn.Close()

	if err := db.Migrate(targetConn); err != nil {
		panic(err)
	}

	rows, err := targetConn.Query("SHOW TABLES")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var tables []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			panic(err)
		}
		tables = append(tables, t)
	}
	fmt.Println("database recreated:", dbName)
	fmt.Println("tables:", strings.Join(tables, ","))
}
