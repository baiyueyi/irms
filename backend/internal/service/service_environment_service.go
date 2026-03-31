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

type ServiceEnvironmentService struct {
	db       *gorm.DB
	q        *query.Query
	auditSvc *AuditService
}

func NewServiceEnvironmentService(st *store.Store) *ServiceEnvironmentService {
	return &ServiceEnvironmentService{db: st.Gorm, q: st.Query, auditSvc: NewAuditService(st.Gorm)}
}

func (s *ServiceEnvironmentService) List(ctx context.Context, serviceID uint64, page int, pageSize int) (list []vo.ServiceEnvironmentVO, total int, err error) {
	qse := s.q.ServiceEnvironment
	items, total64, err := qse.WithContext(ctx).
		Where(qse.ServiceID.Eq(serviceID)).
		Order(qse.ID.Desc()).
		FindByPage((page-1)*pageSize, pageSize)
	if err != nil {
		return nil, 0, err
	}
	out := make([]vo.ServiceEnvironmentVO, 0, len(items))
	for _, it := range items {
		out = append(out, vo.ServiceEnvironmentVO{
			ID:            it.ID,
			ServiceID:     it.ServiceID,
			EnvironmentID: it.EnvironmentID,
			CreatedAt:     it.CreatedAt,
		})
	}
	return out, int(total64), nil
}

func (s *ServiceEnvironmentService) Create(ctx context.Context, actor Actor, serviceID uint64, envID uint64) error {
	now := time.Now()
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		if err := qtx.ServiceEnvironment.WithContext(ctx).Create(&model.ServiceEnvironment{ServiceID: serviceID, EnvironmentID: envID, CreatedAt: now}); err != nil {
			return err
		}
		after := struct {
			ServiceID     uint64 `json:"service_id"`
			EnvironmentID uint64 `json:"environment_id"`
		}{ServiceID: serviceID, EnvironmentID: envID}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "create_service_environment", "service_environments", strconv.FormatUint(serviceID, 10), strconv.FormatUint(serviceID, 10), nil, after)
	})
}

func (s *ServiceEnvironmentService) Delete(ctx context.Context, actor Actor, serviceID uint64, envID uint64) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		qse := qtx.ServiceEnvironment
		if _, err := qse.WithContext(ctx).Where(qse.ServiceID.Eq(serviceID), qse.EnvironmentID.Eq(envID)).Delete(); err != nil {
			return err
		}
		before := struct {
			ServiceID     uint64 `json:"service_id"`
			EnvironmentID uint64 `json:"environment_id"`
		}{ServiceID: serviceID, EnvironmentID: envID}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "delete_service_environment", "service_environments", strconv.FormatUint(serviceID, 10), strconv.FormatUint(serviceID, 10), before, nil)
	})
}
