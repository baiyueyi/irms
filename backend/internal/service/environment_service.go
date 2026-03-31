package service

import (
	"context"
	"strconv"
	"time"

	"irms/backend/internal/dto/request"
	"irms/backend/internal/model"
	"irms/backend/internal/query"
	"irms/backend/internal/store"
	"irms/backend/internal/vo"

	"gorm.io/gorm"
)

type EnvironmentService struct {
	st       *store.Store
	db       *gorm.DB
	q        *query.Query
	auditSvc *AuditService
}

func NewEnvironmentService(st *store.Store) *EnvironmentService {
	return &EnvironmentService{
		st:       st,
		db:       st.Gorm,
		q:        st.Query,
		auditSvc: NewAuditService(st.Gorm),
	}
}

func (s *EnvironmentService) ListPaged(ctx context.Context, keyword string, page int, pageSize int) (list []vo.EnvironmentVO, total int, err error) {
	qe := s.q.Environment
	do := qe.WithContext(ctx)
	if keyword != "" {
		do = do.Where(qe.Name.Like("%" + keyword + "%"))
	}
	items, total64, err := do.Order(qe.ID.Desc()).FindByPage((page-1)*pageSize, pageSize)
	if err != nil {
		return nil, 0, err
	}
	out := make([]vo.EnvironmentVO, 0, len(items))
	for _, it := range items {
		desc := ""
		if it.Description != nil {
			desc = *it.Description
		}
		out = append(out, vo.EnvironmentVO{
			ID:          it.ID,
			Code:        it.Code,
			Name:        it.Name,
			Status:      it.Status,
			Description: desc,
			CreatedAt:   it.CreatedAt,
			UpdatedAt:   it.UpdatedAt,
		})
	}
	return out, int(total64), nil
}

func (s *EnvironmentService) Create(ctx context.Context, actor Actor, req request.EnvironmentCreateRequest) (uint64, error) {
	now := time.Now()
	item := model.Environment{
		Code:        req.Code,
		Name:        req.Name,
		Status:      req.StatusOrDefault(),
		Description: req.DescriptionPtr(),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Create(&item).Error; err != nil {
			return err
		}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "create_environment", "environments", strconv.FormatUint(item.ID, 10), item.Name, nil, req)
	}); err != nil {
		return 0, err
	}
	return item.ID, nil
}

func (s *EnvironmentService) Update(ctx context.Context, actor Actor, id uint64, req request.EnvironmentUpdateRequest) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		stx := s.st.WithTx(tx)
		qe := stx.Query.Environment
		before, err := qe.WithContext(ctx).Where(qe.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if _, err := qe.WithContext(ctx).Where(qe.ID.Eq(id)).UpdateSimple(
			qe.Code.Value(req.Code),
			qe.Name.Value(req.Name),
			qe.Status.Value(req.Status),
			qe.UpdatedAt.Value(time.Now()),
		); err != nil {
			return err
		}
		if _, err := qe.WithContext(ctx).Where(qe.ID.Eq(id)).Update(qe.Description, req.DescriptionPtr()); err != nil {
			return err
		}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "update_environment", "environments", strconv.FormatUint(id, 10), req.Name, before, req)
	})
}

func (s *EnvironmentService) Delete(ctx context.Context, actor Actor, id uint64) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		stx := s.st.WithTx(tx)
		qe := stx.Query.Environment
		before, err := qe.WithContext(ctx).Where(qe.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if _, err := qe.WithContext(ctx).Where(qe.ID.Eq(id)).Delete(); err != nil {
			return err
		}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "delete_environment", "environments", strconv.FormatUint(id, 10), strconv.FormatUint(id, 10), before, nil)
	})
}
