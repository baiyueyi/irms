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

type LocationService struct {
	st       *store.Store
	db       *gorm.DB
	q        *query.Query
	auditSvc *AuditService
}

func NewLocationService(st *store.Store) *LocationService {
	return &LocationService{
		st:       st,
		db:       st.Gorm,
		q:        st.Query,
		auditSvc: NewAuditService(st.Gorm),
	}
}

func (s *LocationService) ListPaged(ctx context.Context, keyword string, page int, pageSize int) (list []vo.LocationVO, total int, err error) {
	ql := s.q.Location
	do := ql.WithContext(ctx)
	if keyword != "" {
		do = do.Where(ql.Name.Like("%" + keyword + "%"))
	}
	items, total64, err := do.Order(ql.ID.Desc()).FindByPage((page-1)*pageSize, pageSize)
	if err != nil {
		return nil, 0, err
	}
	out := make([]vo.LocationVO, 0, len(items))
	for _, it := range items {
		desc := ""
		if it.Description != nil {
			desc = *it.Description
		}
		addr := ""
		if it.Address != nil {
			addr = *it.Address
		}
		out = append(out, vo.LocationVO{
			ID:           it.ID,
			Code:         it.Code,
			Name:         it.Name,
			LocationType: it.LocationType,
			Address:      addr,
			Status:       it.Status,
			Description:  desc,
			CreatedAt:    it.CreatedAt,
			UpdatedAt:    it.UpdatedAt,
		})
	}
	return out, int(total64), nil
}

func (s *LocationService) Create(ctx context.Context, actor Actor, req request.LocationCreateRequest) (uint64, error) {
	now := time.Now()
	item := model.Location{
		Code:         req.Code,
		Name:         req.Name,
		LocationType: req.LocationType,
		Address:      req.AddressPtr(),
		Status:       req.StatusOrDefault(),
		Description:  req.DescriptionPtr(),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Create(&item).Error; err != nil {
			return err
		}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "create_location", "locations", strconv.FormatUint(item.ID, 10), item.Name, nil, req)
	}); err != nil {
		return 0, err
	}
	return item.ID, nil
}

func (s *LocationService) Update(ctx context.Context, actor Actor, id uint64, req request.LocationUpdateRequest) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		stx := s.st.WithTx(tx)
		ql := stx.Query.Location
		before, err := ql.WithContext(ctx).Where(ql.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if _, err := ql.WithContext(ctx).Where(ql.ID.Eq(id)).UpdateSimple(
			ql.Code.Value(req.Code),
			ql.Name.Value(req.Name),
			ql.LocationType.Value(req.LocationType),
			ql.Status.Value(req.Status),
			ql.UpdatedAt.Value(time.Now()),
		); err != nil {
			return err
		}
		if _, err := ql.WithContext(ctx).Where(ql.ID.Eq(id)).Update(ql.Address, req.AddressPtr()); err != nil {
			return err
		}
		if _, err := ql.WithContext(ctx).Where(ql.ID.Eq(id)).Update(ql.Description, req.DescriptionPtr()); err != nil {
			return err
		}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "update_location", "locations", strconv.FormatUint(id, 10), req.Name, before, req)
	})
}

func (s *LocationService) Delete(ctx context.Context, actor Actor, id uint64) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		stx := s.st.WithTx(tx)
		ql := stx.Query.Location
		before, err := ql.WithContext(ctx).Where(ql.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if _, err := ql.WithContext(ctx).Where(ql.ID.Eq(id)).Delete(); err != nil {
			return err
		}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "delete_location", "locations", strconv.FormatUint(id, 10), strconv.FormatUint(id, 10), before, nil)
	})
}
