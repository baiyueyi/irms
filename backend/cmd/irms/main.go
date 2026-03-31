// @title IRMS API
// @version 0.1
// @description IRMS 后端 API（Gin + Gorm + Swagger）
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"irms/backend/internal/bootstrap"
	"irms/backend/internal/config"
	"irms/backend/internal/db"
	"irms/backend/internal/store"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg, err := config.FromEnv()
	if err != nil {
		panic(err)
	}
	conn, err := db.OpenMySQL(cfg)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	if err := db.Migrate(conn); err != nil {
		panic(err)
	}
	if len(os.Args) > 1 && os.Args[1] == "init-superadmin" {
		if err := initSuperAdmin(conn); err != nil {
			panic(err)
		}
		fmt.Println("done")
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "migrate-page-route-path" {
		if err := migratePageRoutePath(conn); err != nil {
			panic(err)
		}
		fmt.Println("done")
		return
	}

	gormDB, err := bootstrap.OpenGormFromSQLConn(conn)
	if err != nil {
		panic(err)
	}
	st, err := store.New(gormDB, conn)
	if err != nil {
		panic(err)
	}
	app := bootstrap.NewApp(cfg, st)

	fmt.Printf("server listening on %s\n", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, app.Engine); err != nil {
		panic(err)
	}
}

func initSuperAdmin(conn *sql.DB) error {
	var cnt int
	if err := conn.QueryRow(`SELECT COUNT(1) FROM users WHERE role='super_admin'`).Scan(&cnt); err != nil {
		return err
	}
	if cnt > 0 {
		return fmt.Errorf("super_admin already exists")
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	fmt.Print("password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	now := time.Now()
	_, err := conn.Exec(`INSERT INTO users(username,password_hash,role,status,must_change_password,created_at,updated_at) VALUES(?,?,?,?,1,?,?)`,
		username, string(hash), "super_admin", "enabled", now, now)
	return err
}

func migratePageRoutePath(conn *sql.DB) error {
	type row struct {
		ID        int64
		RoutePath string
	}
	rows, err := conn.Query(`SELECT id, route_path FROM pages WHERE route_path LIKE '/admin%'`)
	if err != nil {
		return err
	}
	defer rows.Close()
	var items []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.ID, &r.RoutePath); err != nil {
			return err
		}
		items = append(items, r)
	}

	var updated, skipped int
	for _, it := range items {
		newPath := strings.TrimPrefix(it.RoutePath, "/admin")
		if newPath == "" {
			newPath = "/"
		}
		var cnt int
		if err := conn.QueryRow(`SELECT COUNT(1) FROM pages WHERE route_path=? AND id<>?`, newPath, it.ID).Scan(&cnt); err != nil {
			return err
		}
		if cnt > 0 {
			skipped++
			fmt.Printf("skip page id=%d route_path=%s -> %s (conflict)\n", it.ID, it.RoutePath, newPath)
			continue
		}
		if _, err := conn.Exec(`UPDATE pages SET route_path=?, updated_at=? WHERE id=?`, newPath, time.Now(), it.ID); err != nil {
			return err
		}
		updated++
		fmt.Printf("update page id=%d route_path=%s -> %s\n", it.ID, it.RoutePath, newPath)
	}
	fmt.Printf("summary: total=%d updated=%d skipped=%d\n", len(items), updated, skipped)
	return nil
}
