package store

import (
	"database/sql"
	"errors"

	"irms/backend/internal/query"

	"gorm.io/gorm"
)

type Store struct {
	Gorm  *gorm.DB
	SQL   *sql.DB
	Query *query.Query
}

type TxStore struct {
	Gorm  *gorm.DB
	Query *query.Query
}

func New(gormDB *gorm.DB, sqlDB *sql.DB) (*Store, error) {
	if gormDB == nil || sqlDB == nil {
		return nil, errors.New("nil gormDB or sqlDB")
	}
	sqlFromGorm, err := gormDB.DB()
	if err != nil {
		return nil, err
	}
	if sqlFromGorm != sqlDB {
		return nil, errors.New("gorm and sql must share the same *sql.DB pool")
	}
	query.SetDefault(gormDB)
	return &Store{
		Gorm:  gormDB,
		SQL:   sqlDB,
		Query: query.Q,
	}, nil
}

func (s *Store) WithTx(tx *gorm.DB) *TxStore {
	if s == nil || tx == nil {
		return nil
	}
	return &TxStore{
		Gorm:  tx,
		Query: query.Use(tx),
	}
}
