package repository

import (
	"context"

	"gorm.io/gorm"
)

type PermissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

type MyPageResourceRow struct {
	ResourceKey  uint64 `gorm:"column:resource_key"`
	ResourceName string `gorm:"column:resource_name"`
	ResourceType string `gorm:"column:resource_type"`
	RoutePath    string `gorm:"column:route_path"`
	Permission   string `gorm:"column:permission"`
}

func (r *PermissionRepository) ListMyPageResources(ctx context.Context, userID uint64) ([]MyPageResourceRow, error) {
	var rows []MyPageResourceRow
	q := `
SELECT x.resource_key, x.resource_name, x.resource_type, x.route_path, 'page.view' AS permission
FROM (
	SELECT p.id AS resource_key, p.name AS resource_name, 'page' AS resource_type, p.route_path AS route_path
	FROM grants g
	JOIN pages p ON g.object_type='page' AND g.object_id=p.id
	WHERE g.subject_type='user' AND g.subject_id=? AND g.permission_code='page.view' AND p.status='active'
	UNION ALL
	SELECT p.id, p.name, 'page', p.route_path
	FROM user_group_members ugm
	JOIN grants g ON g.subject_type='user_group' AND g.subject_id=ugm.user_group_id AND g.object_type='page'
	JOIN pages p ON p.id=g.object_id
	WHERE ugm.user_id=? AND g.permission_code='page.view' AND p.status='active'
) x
GROUP BY x.resource_key, x.resource_name, x.resource_type, x.route_path
ORDER BY x.route_path ASC, x.resource_key DESC`
	if err := r.db.WithContext(ctx).Raw(q, userID, userID).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
