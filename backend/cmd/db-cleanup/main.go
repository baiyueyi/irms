package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")
	if host == "" || port == "" || user == "" || pass == "" || name == "" {
		panic("missing DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME")
	}
	if !strings.EqualFold(name, "irms") {
		panic("DB_NAME must be irms")
	}
	baseDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=true&loc=Local", user, pass, host, port)
	db, err := sql.Open("mysql", baseDSN)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		panic(err)
	}

	rows, err := db.Query(`SHOW DATABASES`)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var candidates []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			panic(err)
		}
		if strings.HasPrefix(strings.ToLower(n), "irms_") {
			candidates = append(candidates, n)
		}
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}

	fmt.Println("PENDING_DELETE_DATABASES:")
	if len(candidates) == 0 {
		fmt.Println("- <none>")
	} else {
		for _, n := range candidates {
			fmt.Printf("- %s\n", n)
		}
	}

	fmt.Println("DELETE_RESULTS:")
	if len(candidates) == 0 {
		fmt.Println("- <none>")
		return
	}
	for _, n := range candidates {
		stmt := fmt.Sprintf("DROP DATABASE `%s`", strings.ReplaceAll(n, "`", "``"))
		if _, err := db.Exec(stmt); err != nil {
			fmt.Printf("- %s: FAILED (%v)\n", n, err)
			continue
		}
		fmt.Printf("- %s: OK\n", n)
	}
}
