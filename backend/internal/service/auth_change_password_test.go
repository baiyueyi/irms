package service

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"irms/backend/internal/config"
	"irms/backend/internal/db"
	ecode "irms/backend/internal/pkg/errors"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func openTestGorm2(t *testing.T) (*gorm.DB, func()) {
	t.Helper()
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	name := os.Getenv("DB_NAME")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	if host == "" || port == "" || name == "" || user == "" || pass == "" {
		t.Skip("missing DB_HOST/DB_PORT/DB_NAME/DB_USER/DB_PASSWORD")
	}
	targetDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=true&loc=Local", user, pass, host, port, name)
	targetConn, err := sql.Open("mysql", targetDSN)
	if err != nil {
		t.Fatalf("open target mysql: %v", err)
	}
	{
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := targetConn.PingContext(ctx); err != nil {
			_ = targetConn.Close()
			t.Skipf("mysql not reachable: %v", err)
		}
	}
	if err := db.Migrate(targetConn); err != nil {
		_ = targetConn.Close()
		t.Fatalf("migrate: %v", err)
	}
	gormDB, err := gorm.Open(mysql.Open(targetDSN), &gorm.Config{})
	if err != nil {
		_ = targetConn.Close()
		t.Fatalf("open gorm: %v", err)
	}
	cleanup := func() {
		_ = targetConn.Close()
	}
	return gormDB, cleanup
}

func TestChangePassword_OldPasswordMismatchIs401(t *testing.T) {
	gormDB, cleanup := openTestGorm2(t)
	defer cleanup()

	hash, err := bcrypt.GenerateFromPassword([]byte("old-pass"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	now := time.Now()
	username := fmt.Sprintf("u1_%d", now.UnixNano())
	u := struct {
		ID               uint64    `gorm:"column:id"`
		Username         string    `gorm:"column:username"`
		PasswordHash     string    `gorm:"column:password_hash"`
		Role             string    `gorm:"column:role"`
		Status           string    `gorm:"column:status"`
		MustChangePasswd bool      `gorm:"column:must_change_password"`
		CreatedAt        time.Time `gorm:"column:created_at"`
		UpdatedAt        time.Time `gorm:"column:updated_at"`
	}{
		Username:         username,
		PasswordHash:     string(hash),
		Role:             "super_admin",
		Status:           "enabled",
		MustChangePasswd: false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := gormDB.Table("users").Create(&u).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	svc := NewAuthService(config.Config{JWTSecret: "x"}, gormDB)
	actor := Actor{UserID: u.ID, Username: u.Username, Role: "super_admin", IP: "127.0.0.1"}
	err = svc.ChangePassword(context.Background(), actor, u.ID, "wrong-old", "new-pass")
	if err == nil {
		t.Fatalf("expected error")
	}
	ae, ok := err.(*ecode.AppError)
	if !ok {
		t.Fatalf("expected *AppError, got %T: %v", err, err)
	}
	if ae.Status != 401 || ae.Code != ecode.CodeInvalidCredentials {
		t.Fatalf("expected 401 INVALID_CREDENTIALS, got %d %s", ae.Status, ae.Code)
	}
}

func TestChangePassword_UserNotFoundNotMappedToOldPasswordMismatch(t *testing.T) {
	gormDB, cleanup := openTestGorm2(t)
	defer cleanup()

	svc := NewAuthService(config.Config{JWTSecret: "x"}, gormDB)
	actor := Actor{UserID: 999, Username: "u", Role: "super_admin", IP: "127.0.0.1"}
	err := svc.ChangePassword(context.Background(), actor, 999, "old", "new")
	if err == nil {
		t.Fatalf("expected error")
	}
	if _, ok := err.(*ecode.AppError); ok {
		t.Fatalf("expected non-AppError not found/db error, got AppError")
	}
}
