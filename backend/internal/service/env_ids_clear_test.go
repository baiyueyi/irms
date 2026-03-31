package service

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"irms/backend/internal/db"
	"irms/backend/internal/dto/request"
	"irms/backend/internal/model"
	"irms/backend/internal/store"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func openTestGorm(t *testing.T) (*gorm.DB, func()) {
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

func seedEnvAndLocation(t *testing.T, gormDB *gorm.DB) (env1, env2 uint64, locID uint64) {
	t.Helper()
	now := time.Now()
	suffix := fmt.Sprintf("%d", now.UnixNano())
	e1 := model.Environment{Code: "dev_" + suffix, Name: "Dev " + suffix, Status: "active", CreatedAt: now, UpdatedAt: now}
	e2 := model.Environment{Code: "prod_" + suffix, Name: "Prod " + suffix, Status: "active", CreatedAt: now, UpdatedAt: now}
	loc := model.Location{Code: "bj_" + suffix, Name: "Beijing " + suffix, LocationType: "idc", Status: "active", CreatedAt: now, UpdatedAt: now}
	if err := gormDB.Create(&e1).Error; err != nil {
		t.Fatalf("create env1: %v", err)
	}
	if err := gormDB.Create(&e2).Error; err != nil {
		t.Fatalf("create env2: %v", err)
	}
	if err := gormDB.Create(&loc).Error; err != nil {
		t.Fatalf("create location: %v", err)
	}
	return e1.ID, e2.ID, loc.ID
}

func seedUser(t *testing.T, gormDB *gorm.DB) uint64 {
	t.Helper()
	now := time.Now()
	u := model.User{
		Username:           fmt.Sprintf("tester_%d", now.UnixNano()),
		PasswordHash:       "x",
		Role:               "super_admin",
		Status:             "enabled",
		MustChangePassword: false,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := gormDB.Create(&u).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return u.ID
}

func TestHostUpdateEnvironmentIDsClear(t *testing.T) {
	gormDB, cleanup := openTestGorm(t)
	defer cleanup()

	st, err := store.New(gormDB, mustSQLDB(t, gormDB))
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	userID := seedUser(t, gormDB)
	env1, env2, locID := seedEnvAndLocation(t, gormDB)

	ctlCtx := context.Background()
	actor := Actor{UserID: userID, Username: "tester", Role: "super_admin", IP: "127.0.0.1"}
	hs := NewHostService(st)

	createEnvs := []uint64{env1, env2}
	hostID, err := hs.Create(ctlCtx, actor, request.HostCreateRequest{
		Name:           "host-a",
		Hostname:       "host-a.internal",
		PrimaryAddress: "10.0.0.1",
		ProviderKind:   "vm",
		OSType:         "linux",
		Status:         "active",
		LocationID:     &locID,
		EnvironmentIDs: &createEnvs,
	})
	if err != nil {
		t.Fatalf("create host: %v", err)
	}

	empty := []uint64{}
	if err := hs.Update(ctlCtx, actor, hostID, request.HostUpdateRequest{
		Name:           "host-a",
		Hostname:       "host-a.internal",
		PrimaryAddress: "10.0.0.1",
		ProviderKind:   "vm",
		OSType:         "linux",
		Status:         "active",
		LocationID:     &locID,
		Description:    "",
		EnvironmentIDs: &empty,
	}); err != nil {
		t.Fatalf("update host: %v", err)
	}

	got, err := hs.GetByID(ctlCtx, hostID)
	if err != nil {
		t.Fatalf("get host: %v", err)
	}
	if len(got.EnvironmentIDs) != 0 {
		t.Fatalf("expected env cleared, got=%v", got.EnvironmentIDs)
	}
	if got.EnvironmentSource != "none" {
		t.Fatalf("expected source none, got=%s", got.EnvironmentSource)
	}
}

func TestServiceUpdateEnvironmentIDsClearFallbackToHost(t *testing.T) {
	gormDB, cleanup := openTestGorm(t)
	defer cleanup()

	st, err := store.New(gormDB, mustSQLDB(t, gormDB))
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	userID := seedUser(t, gormDB)
	env1, env2, locID := seedEnvAndLocation(t, gormDB)

	ctlCtx := context.Background()
	actor := Actor{UserID: userID, Username: "tester", Role: "super_admin", IP: "127.0.0.1"}
	hs := NewHostService(st)
	ss := NewServiceService(st)

	hostEnvs := []uint64{env1, env2}
	hostID, err := hs.Create(ctlCtx, actor, request.HostCreateRequest{
		Name:           "host-a",
		Hostname:       "host-a.internal",
		PrimaryAddress: "10.0.0.1",
		ProviderKind:   "vm",
		OSType:         "linux",
		Status:         "active",
		LocationID:     &locID,
		EnvironmentIDs: &hostEnvs,
	})
	if err != nil {
		t.Fatalf("create host: %v", err)
	}

	svcEnvs := []uint64{env2}
	svcID, err := ss.Create(ctlCtx, actor, request.ServiceCreateRequest{
		Name:                 "svc-a",
		ServiceKind:          "app",
		HostID:               &hostID,
		EndpointOrIdentifier: "svc-a.internal",
		Port:                 func() *int { v := 8080; return &v }(),
		Protocol:             "http",
		Status:               "active",
		Description:          "",
		EnvironmentIDs:       &svcEnvs,
	})
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	empty := []uint64{}
	if err := ss.Update(ctlCtx, actor, svcID, request.ServiceUpdateRequest{
		Name:                 "svc-a",
		ServiceKind:          "app",
		HostID:               &hostID,
		EndpointOrIdentifier: "svc-a.internal",
		Port:                 func() *int { v := 8080; return &v }(),
		Protocol:             "http",
		Status:               "active",
		Description:          "",
		EnvironmentIDs:       &empty,
	}); err != nil {
		t.Fatalf("update service: %v", err)
	}

	got, err := ss.GetByID(ctlCtx, svcID)
	if err != nil {
		t.Fatalf("get service: %v", err)
	}
	if got.EnvironmentSource != "host_inherited" {
		t.Fatalf("expected source host_inherited, got=%s", got.EnvironmentSource)
	}
	if len(got.EnvironmentIDs) != 2 {
		t.Fatalf("expected inherited env ids, got=%v", got.EnvironmentIDs)
	}
}

func mustSQLDB(t *testing.T, gormDB *gorm.DB) *sql.DB {
	t.Helper()
	sqlDB, err := gormDB.DB()
	if err != nil {
		t.Fatalf("gorm.DB(): %v", err)
	}
	return sqlDB
}
