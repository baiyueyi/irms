package service

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"irms/backend/internal/dto/request"
	"irms/backend/internal/model"
	"irms/backend/internal/pkg/dbutil"
	ecode "irms/backend/internal/pkg/errors"
	"irms/backend/internal/query"
	"irms/backend/internal/store"
	"irms/backend/internal/vo"

	"gorm.io/gorm"
)

type ResourceService struct {
	db       *gorm.DB
	q        *query.Query
	auditSvc *AuditService
}

// NewResourceService 保留用于 compatibility storage。
// Deprecated: 请勿在 resources 三表上扩展新业务能力。
func NewResourceService(st *store.Store) *ResourceService {
	return &ResourceService{
		db:       st.Gorm,
		q:        st.Query,
		auditSvc: NewAuditService(st.Gorm),
	}
}

// Deprecated: resources 列表接口仅用于 compatibility storage 运维与历史数据维护。
func (s *ResourceService) ListPaged(ctx context.Context, f ResourceListFilter, page int, pageSize int) ([]vo.ResourceVO, int, error) {
	qr := s.q.Resource
	do := qr.WithContext(ctx)
	if v := strings.TrimSpace(f.Keyword); v != "" {
		do = do.Where(qr.Name.Like("%" + v + "%"))
	}
	if v := strings.TrimSpace(f.Status); v != "" {
		do = do.Where(qr.Status.Eq(v))
	}
	if v := strings.TrimSpace(f.Type); v != "" {
		do = do.Where(qr.Type.Eq(v))
	} else {
		do = do.Where(qr.Type.In("host", "service"))
	}
	items, total64, err := do.Order(qr.Key.Desc()).FindByPage((page-1)*pageSize, pageSize)
	if err != nil {
		return nil, 0, err
	}
	out := make([]vo.ResourceVO, 0, len(items))
	for _, it := range items {
		out = append(out, toResourceVO(*it))
	}
	return out, int(total64), nil
}

// Deprecated: resources 创建接口仅用于 compatibility storage。
func (s *ResourceService) Create(ctx context.Context, actor Actor, req request.ResourceCreateRequest) (uint64, error) {
	if !validResourceType(req.Type) {
		return 0, ecode.NewAppError(http.StatusBadRequest, ecode.CodeInvalidArgument, "invalid resource type", nil)
	}
	if req.Type == "host" && strings.TrimSpace(req.Address) == "" {
		return 0, ecode.NewAppError(http.StatusBadRequest, ecode.CodeInvalidArgument, "address required for host", nil)
	}
	if req.Type == "service" && strings.TrimSpace(req.ServiceIdentifier) == "" {
		return 0, ecode.NewAppError(http.StatusBadRequest, ecode.CodeInvalidArgument, "service_identifier required for service", nil)
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "active"
	}
	now := time.Now()
	it := model.Resource{
		Name:      strings.TrimSpace(req.Name),
		Type:      req.Type,
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if v := strings.TrimSpace(req.Address); v != "" {
		it.Address = &v
	}
	if v := strings.TrimSpace(req.ServiceIdentifier); v != "" {
		it.ServiceIdentifier = &v
	}
	if v := strings.TrimSpace(req.RoutePath); v != "" {
		it.RoutePath = &v
	}
	if v := strings.TrimSpace(req.Description); v != "" {
		it.Description = &v
	}
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		if err := qtx.Resource.WithContext(ctx).Create(&it); err != nil {
			if dbutil.IsDuplicate(err) {
				return ecode.NewAppError(http.StatusConflict, ecode.CodeConflict, "resource exists", nil)
			}
			return err
		}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "create_resource", "resources", strconv.FormatUint(it.Key, 10), it.Name, nil, req)
	}); err != nil {
		return 0, err
	}
	return it.Key, nil
}

// Deprecated: resources 更新接口仅用于 compatibility storage。
func (s *ResourceService) Update(ctx context.Context, actor Actor, key uint64, req request.ResourceUpdateRequest) error {
	if !validResourceType(req.Type) {
		return ecode.NewAppError(http.StatusBadRequest, ecode.CodeInvalidArgument, "invalid resource type", nil)
	}
	if req.Type == "host" && strings.TrimSpace(req.Address) == "" {
		return ecode.NewAppError(http.StatusBadRequest, ecode.CodeInvalidArgument, "address required for host", nil)
	}
	if req.Type == "service" && strings.TrimSpace(req.ServiceIdentifier) == "" {
		return ecode.NewAppError(http.StatusBadRequest, ecode.CodeInvalidArgument, "service_identifier required for service", nil)
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		qr := qtx.Resource
		before, err := qr.WithContext(ctx).Where(qr.Key.Eq(key)).First()
		if err != nil {
			return err
		}
		status := strings.TrimSpace(req.Status)
		if status == "" {
			status = "active"
		}
		now := time.Now()
		if _, err := qr.WithContext(ctx).Where(qr.Key.Eq(key)).UpdateSimple(
			qr.Name.Value(strings.TrimSpace(req.Name)),
			qr.Type.Value(req.Type),
			qr.Status.Value(status),
			qr.UpdatedAt.Value(now),
		); err != nil {
			return err
		}
		if _, err := qr.WithContext(ctx).Where(qr.Key.Eq(key)).Update(qr.Address, nullOrString(strings.TrimSpace(req.Address))); err != nil {
			return err
		}
		if _, err := qr.WithContext(ctx).Where(qr.Key.Eq(key)).Update(qr.ServiceIdentifier, nullOrString(strings.TrimSpace(req.ServiceIdentifier))); err != nil {
			return err
		}
		if _, err := qr.WithContext(ctx).Where(qr.Key.Eq(key)).Update(qr.RoutePath, nullOrString(strings.TrimSpace(req.RoutePath))); err != nil {
			return err
		}
		if _, err := qr.WithContext(ctx).Where(qr.Key.Eq(key)).Update(qr.Description, nullOrString(strings.TrimSpace(req.Description))); err != nil {
			return err
		}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "update_resource", "resources", strconv.FormatUint(key, 10), req.Name, before, req)
	})
}

// Deprecated: resources 删除接口仅用于 compatibility storage。
func (s *ResourceService) Delete(ctx context.Context, actor Actor, key uint64) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		qr := qtx.Resource
		before, err := qr.WithContext(ctx).Where(qr.Key.Eq(key)).First()
		if err != nil {
			return err
		}
		if _, err := qr.WithContext(ctx).Where(qr.Key.Eq(key)).Delete(); err != nil {
			return err
		}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "delete_resource", "resources", strconv.FormatUint(key, 10), before.Name, before, nil)
	})
}

func validResourceType(t string) bool {
	return t == "host" || t == "service"
}

func toResourceVO(it model.Resource) vo.ResourceVO {
	out := vo.ResourceVO{
		Key:       it.Key,
		Name:      it.Name,
		Type:      it.Type,
		Status:    it.Status,
		CreatedAt: it.CreatedAt,
		UpdatedAt: it.UpdatedAt,
	}
	if it.Address != nil {
		out.Address = *it.Address
	}
	if it.ServiceIdentifier != nil {
		out.ServiceIdentifier = *it.ServiceIdentifier
	}
	if it.RoutePath != nil {
		out.RoutePath = *it.RoutePath
	}
	if it.Description != nil {
		out.Description = *it.Description
	}
	return out
}
