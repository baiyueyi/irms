package service

import (
	"context"
	"strconv"
	"time"

	"irms/backend/internal/model"
	"irms/backend/internal/query"
	"irms/backend/internal/store"
	"irms/backend/internal/vo"

	"gorm.io/gorm"
)

type HostEnvironmentService struct {
	db       *gorm.DB
	q        *query.Query
	auditSvc *AuditService
}

func NewHostEnvironmentService(st *store.Store) *HostEnvironmentService {
	return &HostEnvironmentService{db: st.Gorm, q: st.Query, auditSvc: NewAuditService(st.Gorm)}
}

func (s *HostEnvironmentService) List(ctx context.Context, hostID uint64, page int, pageSize int) (list []vo.HostEnvironmentVO, total int, err error) {
	qhe := s.q.HostEnvironment
	items, total64, err := qhe.WithContext(ctx).
		Where(qhe.HostID.Eq(hostID)).
		Order(qhe.ID.Desc()).
		FindByPage((page-1)*pageSize, pageSize)
	if err != nil {
		return nil, 0, err
	}
	out := make([]vo.HostEnvironmentVO, 0, len(items))
	for _, it := range items {
		out = append(out, vo.HostEnvironmentVO{
			ID:            it.ID,
			HostID:        it.HostID,
			EnvironmentID: it.EnvironmentID,
			CreatedAt:     it.CreatedAt,
		})
	}
	return out, int(total64), nil
}

func (s *HostEnvironmentService) Create(ctx context.Context, actor Actor, hostID uint64, envID uint64) error {
	now := time.Now()
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		if err := qtx.HostEnvironment.WithContext(ctx).Create(&model.HostEnvironment{HostID: hostID, EnvironmentID: envID, CreatedAt: now}); err != nil {
			return err
		}
		after := struct {
			HostID        uint64 `json:"host_id"`
			EnvironmentID uint64 `json:"environment_id"`
		}{HostID: hostID, EnvironmentID: envID}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "create_host_environment", "host_environments", strconv.FormatUint(hostID, 10), strconv.FormatUint(hostID, 10), nil, after)
	})
}

func (s *HostEnvironmentService) Delete(ctx context.Context, actor Actor, hostID uint64, envID uint64) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		qhe := qtx.HostEnvironment
		if _, err := qhe.WithContext(ctx).Where(qhe.HostID.Eq(hostID), qhe.EnvironmentID.Eq(envID)).Delete(); err != nil {
			return err
		}
		before := struct {
			HostID        uint64 `json:"host_id"`
			EnvironmentID uint64 `json:"environment_id"`
		}{HostID: hostID, EnvironmentID: envID}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "delete_host_environment", "host_environments", strconv.FormatUint(hostID, 10), strconv.FormatUint(hostID, 10), before, nil)
	})
}
