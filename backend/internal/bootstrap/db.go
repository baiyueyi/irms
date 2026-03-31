package bootstrap

import (
	"database/sql"

	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func OpenGormFromSQLConn(sqlConn *sql.DB) (*gorm.DB, error) {
	return gorm.Open(gormmysql.New(gormmysql.Config{
		Conn: sqlConn,
	}), &gorm.Config{})
}
