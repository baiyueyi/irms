package repository

import (
	"context"
	"errors"
	"time"

	"irms/backend/internal/model"
	"irms/backend/internal/query"

	"gorm.io/gorm"
)

type GrantRepository struct {
	db *gorm.DB
}

func NewGrantRepository(db *gorm.DB) *GrantRepository {
	return &GrantRepository{db: db}
}

type GrantListRow struct {
	ID          uint64    `gorm:"column:id"`
	SubjectType string    `gorm:"column:subject_type"`
	SubjectID   uint64    `gorm:"column:subject_id"`
	ObjectType  string    `gorm:"column:object_type"`
	ObjectID    uint64    `gorm:"column:object_id"`
	Permission  string    `gorm:"column:permission"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

type GrantKey struct {
	SubjectType    string
	SubjectID      uint64
	ObjectType     string
	ObjectID       uint64
	PermissionCode string
}

type GrantQueryFilter struct {
	Keyword     string
	SubjectType string
	SubjectID   *uint64
	ObjectType  string
	ObjectID    *uint64
	Permission  string
}

func (r *GrantRepository) CountByFilter(ctx context.Context, f GrantQueryFilter) (int64, error) {
	where, args := buildGrantListWhere(f)
	var total64 int64
	if err := r.db.WithContext(ctx).Raw(`SELECT COUNT(1) FROM grants`+where, args...).Scan(&total64).Error; err != nil {
		return 0, err
	}
	return total64, nil
}

func (r *GrantRepository) ListByFilter(ctx context.Context, f GrantQueryFilter, limit int, offset int) ([]GrantListRow, error) {
	where, args := buildGrantListWhere(f)
	qArgs := append([]interface{}{}, args...)
	qArgs = append(qArgs, limit, offset)
	var rows []GrantListRow
	if err := r.db.WithContext(ctx).Raw(
		`SELECT id,subject_type,subject_id,object_type,object_id,permission_code AS permission,created_at,updated_at FROM grants`+where+` ORDER BY id DESC LIMIT ? OFFSET ?`,
		qArgs...,
	).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *GrantRepository) FindByID(ctx context.Context, tx *gorm.DB, id uint64) (*model.Grant, error) {
	qtx := query.Use(dbOrTx(r.db, tx))
	qg := qtx.Grant
	return qg.WithContext(ctx).Where(qg.ID.Eq(id)).First()
}

func (r *GrantRepository) UpdatePermissionByID(ctx context.Context, tx *gorm.DB, id uint64, permission string) error {
	if tx == nil {
		return gorm.ErrInvalidTransaction
	}
	qtx := query.Use(tx)
	qg := qtx.Grant
	if _, err := qg.WithContext(ctx).
		Where(qg.ID.Eq(id)).
		UpdateSimple(qg.PermissionCode.Value(permission)); err != nil {
		return err
	}
	return nil
}

func (r *GrantRepository) Upsert(ctx context.Context, tx *gorm.DB, key GrantKey) (uint64, error) {
	if tx == nil {
		return 0, gorm.ErrInvalidTransaction
	}
	qtx := query.Use(tx)
	qg := qtx.Grant
	existing, err := qg.WithContext(ctx).
		Where(
			qg.SubjectType.Eq(key.SubjectType),
			qg.SubjectID.Eq(key.SubjectID),
			qg.ObjectType.Eq(key.ObjectType),
			qg.ObjectID.Eq(key.ObjectID),
			qg.PermissionCode.Eq(key.PermissionCode),
		).
		First()
	if err == nil {
		return existing.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}
	it := model.Grant{
		SubjectType:    key.SubjectType,
		SubjectID:      key.SubjectID,
		ObjectType:     key.ObjectType,
		ObjectID:       key.ObjectID,
		PermissionCode: key.PermissionCode,
	}
	if err := qg.WithContext(ctx).Create(&it); err != nil {
		return 0, err
	}
	return it.ID, nil
}

func (r *GrantRepository) SubjectHasAnyPermissionCode(ctx context.Context, subjectType string, subjectID uint64, objectType string, objectID uint64, codes []string) (bool, error) {
	if len(codes) == 0 {
		return false, nil
	}
	var cnt int64
	if err := r.db.WithContext(ctx).Raw(`
SELECT COUNT(1)
FROM grants g
WHERE
(
	(g.subject_type=? AND g.subject_id=?)
	OR
	(
		?='user' AND g.subject_type='user_group' AND EXISTS(
			SELECT 1 FROM user_group_members ugm WHERE ugm.user_id=? AND ugm.user_group_id=g.subject_id
		)
	)
)
AND
(
	(g.object_type=? AND g.object_id=?)
	OR
	(
		?='host' AND g.object_type='host_group' AND EXISTS(
			SELECT 1
			FROM resource_groups rg
			JOIN resource_group_members rgm ON rgm.resource_group_id=rg.id
			WHERE rg.id=g.object_id AND rg.type='host' AND rgm.resource_key=?
		)
	)
	OR
	(
		?='service' AND g.object_type='service_group' AND EXISTS(
			SELECT 1
			FROM resource_groups rg
			JOIN resource_group_members rgm ON rgm.resource_group_id=rg.id
			WHERE rg.id=g.object_id AND rg.type='service' AND rgm.resource_key=?
		)
	)
)
AND g.permission_code IN ?
`,
		subjectType, subjectID,
		subjectType, subjectID,
		objectType, objectID,
		objectType, objectID,
		objectType, objectID,
		codes,
	).Scan(&cnt).Error; err != nil {
		return false, err
	}
	return cnt > 0, nil
}

func buildGrantListWhere(f GrantQueryFilter) (string, []interface{}) {
	where := " WHERE 1=1 "
	args := make([]interface{}, 0, 12)
	if f.Keyword != "" {
		kw := "%" + f.Keyword + "%"
		where += ` AND (
			(subject_type='user' AND EXISTS(SELECT 1 FROM users u WHERE u.id=grants.subject_id AND u.username LIKE ?))
			OR (subject_type='user_group' AND EXISTS(SELECT 1 FROM user_groups ug WHERE ug.id=grants.subject_id AND ug.name LIKE ?))
			OR (object_type='page' AND EXISTS(SELECT 1 FROM pages p WHERE p.id=grants.object_id AND p.name LIKE ?))
			OR (object_type='host' AND EXISTS(SELECT 1 FROM hosts h WHERE h.id=grants.object_id AND h.name LIKE ?))
			OR (object_type='service' AND EXISTS(SELECT 1 FROM services ss WHERE ss.id=grants.object_id AND ss.name LIKE ?))
			OR (object_type='host_credential' AND EXISTS(SELECT 1 FROM hosts h2 WHERE h2.id=grants.object_id AND h2.name LIKE ?))
			OR (object_type='service_credential' AND EXISTS(SELECT 1 FROM services s2 WHERE s2.id=grants.object_id AND s2.name LIKE ?))
			OR (object_type='host_group' AND EXISTS(SELECT 1 FROM resource_groups rg WHERE rg.id=grants.object_id AND rg.type='host' AND rg.name LIKE ?))
			OR (object_type='service_group' AND EXISTS(SELECT 1 FROM resource_groups rg WHERE rg.id=grants.object_id AND rg.type='service' AND rg.name LIKE ?))
		) `
		args = append(args, kw, kw, kw, kw, kw, kw, kw, kw, kw)
	}
	if f.SubjectType != "" {
		where += " AND subject_type=? "
		args = append(args, f.SubjectType)
	}
	if f.SubjectID != nil && *f.SubjectID > 0 {
		where += " AND subject_id=? "
		args = append(args, *f.SubjectID)
	}
	if f.ObjectType != "" {
		where += " AND object_type=? "
		args = append(args, f.ObjectType)
	}
	if f.ObjectID != nil && *f.ObjectID > 0 {
		where += " AND object_id=? "
		args = append(args, *f.ObjectID)
	}
	if f.Permission != "" {
		where += " AND permission_code=? "
		args = append(args, f.Permission)
	}
	return where, args
}

func dbOrTx(db *gorm.DB, tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return db
}
