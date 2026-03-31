package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func main() {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")
	if host == "" || port == "" || user == "" || pass == "" || name == "" {
		panic("missing DB_* envs: DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME")
	}
	if !strings.EqualFold(name, "irms") {
		panic("DB_NAME must be irms")
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=true&loc=Local", user, pass, host, port, name)
	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer sqlDB.Close()

	gormDB, err := gorm.Open(gormmysql.New(gormmysql.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	cfg := gen.Config{
		OutPath:       "internal/query",
		ModelPkgPath:  "model",
		FieldSignable: true,
		FieldNullable: true,
		Mode:          gen.WithDefaultQuery | gen.WithQueryInterface,
	}
	cfg.WithDataTypeMap(map[string]func(columnType gorm.ColumnType) (dataType string){
		"bigint": func(columnType gorm.ColumnType) (dataType string) { return "uint64" },
	})
	cfg.WithJSONTagNameStrategy(func(columnName string) (tagContent string) { return "" })
	g := gen.NewGenerator(cfg)
	g.UseDB(gormDB)

	auditLog := g.GenerateModel("audit_logs")
	environment := g.GenerateModel("environments")
	grant := g.GenerateModel("grants")
	permissionDefinition := g.GenerateModel("permission_definitions")
	hostCredential := g.GenerateModel("host_credentials")
	hostEnvironment := g.GenerateModel("host_environments")
	hostModel := g.GenerateModel("hosts")
	location := g.GenerateModel("locations")
	page := g.GenerateModel("pages")
	resource := g.GenerateModel("resources")
	resourceGroup := g.GenerateModel("resource_groups")
	resourceGroupMember := g.GenerateModel("resource_group_members")
	serviceCredential := g.GenerateModel("service_credentials")
	serviceEnvironment := g.GenerateModel("service_environments")
	service := g.GenerateModel("services")
	userModel := g.GenerateModel("users")
	userGroup := g.GenerateModel("user_groups")
	userGroupMember := g.GenerateModel("user_group_members")

	g.ApplyBasic(
		auditLog,
		environment,
		grant,
		permissionDefinition,
		hostCredential,
		hostEnvironment,
		hostModel,
		location,
		page,
		resource,
		resourceGroup,
		resourceGroupMember,
		service,
		serviceCredential,
		serviceEnvironment,
		userModel,
		userGroup,
		userGroupMember,
	)
	g.Execute()
}
