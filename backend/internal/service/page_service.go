package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	"irms/backend/internal/dto/request"
	"irms/backend/internal/model"
	"irms/backend/internal/pkg/convert"
	"irms/backend/internal/query"
	"irms/backend/internal/store"
	"irms/backend/internal/vo"

	"gorm.io/gorm"
)

type PageService struct {
	db       *gorm.DB
	q        *query.Query
	auditSvc *AuditService
}

func NewPageService(st *store.Store) *PageService {
	return &PageService{db: st.Gorm, q: st.Query, auditSvc: NewAuditService(st.Gorm)}
}

func (s *PageService) ListPaged(ctx context.Context, keyword string, status string, page int, pageSize int) ([]vo.PageVO, int, error) {
	qp := s.q.Page
	do := qp.WithContext(ctx)
	if v := strings.TrimSpace(keyword); v != "" {
		kw := "%" + v + "%"
		do = do.Where(qp.Name.Like(kw)).Or(qp.RoutePath.Like(kw))
	}
	if v := strings.TrimSpace(status); v != "" {
		do = do.Where(qp.Status.Eq(v))
	}
	items, total64, err := do.Order(qp.ID.Desc()).FindByPage((page-1)*pageSize, pageSize)
	if err != nil {
		return nil, 0, err
	}

	grantCountByPageID := map[uint64]int{}
	if len(items) > 0 {
		pageIDs := make([]uint64, 0, len(items))
		for _, it := range items {
			pageIDs = append(pageIDs, it.ID)
		}
		qg := s.q.Grant
		grants, err := qg.WithContext(ctx).
			Select(qg.ObjectID).
			Where(qg.ObjectType.Eq("page"), qg.ObjectID.In(pageIDs...)).
			Find()
		if err != nil {
			return nil, 0, err
		}
		for _, g := range grants {
			grantCountByPageID[g.ObjectID]++
		}
	}

	out := make([]vo.PageVO, 0, len(items))
	for _, r := range items {
		desc := ""
		if r.Description != nil {
			desc = *r.Description
		}
		out = append(out, vo.PageVO{
			ID:          r.ID,
			Name:        r.Name,
			RoutePath:   r.RoutePath,
			Source:      r.Source,
			Status:      r.Status,
			Description: desc,
			GrantCount:  grantCountByPageID[r.ID],
			CreatedAt:   r.CreatedAt,
			UpdatedAt:   r.UpdatedAt,
		})
	}
	return out, int(total64), nil
}

func (s *PageService) Create(ctx context.Context, actor Actor, req request.PageCreateRequest) (uint64, error) {
	now := time.Now()
	source := normalizePageSource(req.Source, "manual")
	status := req.Status
	if strings.TrimSpace(status) == "" {
		status = "active"
	}
	routePath := convert.CanonicalPageRoutePath(strings.TrimSpace(req.RoutePath))
	item := model.Page{
		Name:      req.Name,
		RoutePath: routePath,
		Source:    source,
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if strings.TrimSpace(req.Description) != "" {
		v := req.Description
		item.Description = &v
	}
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		if err := qtx.Page.WithContext(ctx).Create(&item); err != nil {
			return err
		}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "create_page", "pages", strconv.FormatUint(item.ID, 10), req.Name, nil, req)
	}); err != nil {
		return 0, err
	}
	return item.ID, nil
}

func (s *PageService) Update(ctx context.Context, actor Actor, id uint64, req request.PageUpdateRequest) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		qp := qtx.Page
		before, err := qp.WithContext(ctx).Where(qp.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		now := time.Now()
		if _, err := qp.WithContext(ctx).Where(qp.ID.Eq(id)).UpdateSimple(
			qp.Name.Value(req.Name),
			qp.RoutePath.Value(convert.CanonicalPageRoutePath(strings.TrimSpace(req.RoutePath))),
			qp.Source.Value(normalizePageSource(req.Source, "manual")),
			qp.Status.Value(req.Status),
			qp.UpdatedAt.Value(now),
		); err != nil {
			return err
		}
		if _, err := qp.WithContext(ctx).Where(qp.ID.Eq(id)).Update(qp.Description, nullOrString(strings.TrimSpace(req.Description))); err != nil {
			return err
		}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "update_page", "pages", strconv.FormatUint(id, 10), req.Name, *before, req)
	})
}

func (s *PageService) Delete(ctx context.Context, actor Actor, id uint64) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		qp := qtx.Page
		before, err := qp.WithContext(ctx).Where(qp.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if _, err := qp.WithContext(ctx).Where(qp.ID.Eq(id)).Delete(); err != nil {
			return err
		}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "delete_page", "pages", strconv.FormatUint(id, 10), before.Name, *before, nil)
	})
}

func (s *PageService) Sync(ctx context.Context, actor Actor, req request.PageSyncRequest) (vo.PageSyncVO, error) {
	candidates := map[string]request.PageSyncRouteItem{}
	for _, item := range req.Routes {
		rp := convert.CanonicalPageRoutePath(strings.TrimSpace(item.RoutePath))
		if rp == "" || !strings.HasPrefix(rp, "/") {
			continue
		}
		item.RoutePath = rp
		item.Source = normalizePageSource(item.Source, "router")
		item.Name = strings.TrimSpace(item.Name)
		candidates[rp] = item
	}

	routePaths := make([]string, 0, len(candidates))
	for rp := range candidates {
		routePaths = append(routePaths, rp)
	}

	var existing []*model.Page
	var err error
	if len(routePaths) > 0 {
		qp := s.q.Page
		existing, err = qp.WithContext(ctx).Where(qp.RoutePath.In(routePaths...)).Find()
		if err != nil {
			return vo.PageSyncVO{}, err
		}
	}
	existingByPath := map[string]*model.Page{}
	for _, p := range existing {
		existingByPath[p.RoutePath] = p
	}

	newRoutes := []vo.PageSyncRouteVO{}
	existingRoutes := []vo.PageSyncRouteVO{}
	changedRoutes := []vo.PageSyncRouteVO{}
	for rp, cand := range candidates {
		if p, ok := existingByPath[rp]; !ok {
			newRoutes = append(newRoutes, vo.PageSyncRouteVO{RoutePath: rp, Name: cand.Name, Source: cand.Source, Status: "active"})
		} else {
			existingRoutes = append(existingRoutes, vo.PageSyncRouteVO{RoutePath: rp, Name: p.Name, Source: p.Source, Status: p.Status})
			if p.Name != cand.Name || p.Source != cand.Source || p.Status != "active" {
				changedRoutes = append(changedRoutes, vo.PageSyncRouteVO{RoutePath: rp, Name: cand.Name, Source: cand.Source, Status: "active"})
			}
		}
	}

	qp := s.q.Page
	retiredDO := qp.WithContext(ctx).Where(qp.Source.Neq("manual"), qp.Status.Neq("inactive"))
	if len(routePaths) > 0 {
		retiredDO = retiredDO.Where(qp.RoutePath.NotIn(routePaths...))
	}
	retired, err := retiredDO.Find()
	if err != nil {
		return vo.PageSyncVO{}, err
	}
	retiredViews := make([]vo.PageSyncRouteVO, 0, len(retired))
	for _, p := range retired {
		retiredViews = append(retiredViews, vo.PageSyncRouteVO{RoutePath: p.RoutePath, Name: p.Name, Source: p.Source, Status: "inactive"})
	}

	summary := vo.PageSyncSummaryVO{
		DryRun:     req.DryRun,
		InputTotal: len(req.Routes),
		Created:    len(newRoutes),
		Updated:    len(changedRoutes),
		Unchanged:  len(existingRoutes) - len(changedRoutes),
		Retired:    len(retiredViews),
	}
	resp := vo.PageSyncVO{
		Summary:        summary,
		NewRoutes:      newRoutes,
		ExistingRoutes: existingRoutes,
		ChangedRoutes:  changedRoutes,
		RetiredRoutes:  retiredViews,
	}

	if req.DryRun {
		return resp, nil
	}

	now := time.Now()
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		qp := qtx.Page
		for _, r := range newRoutes {
			p := model.Page{
				Name:      r.Name,
				RoutePath: r.RoutePath,
				Source:    r.Source,
				Status:    "active",
				CreatedAt: now,
				UpdatedAt: now,
			}
			if err := qp.WithContext(ctx).Create(&p); err != nil {
				return err
			}
		}
		for _, r := range changedRoutes {
			if _, err := qp.WithContext(ctx).Where(qp.RoutePath.Eq(r.RoutePath)).UpdateSimple(
				qp.Name.Value(r.Name),
				qp.Source.Value(r.Source),
				qp.Status.Value("active"),
				qp.UpdatedAt.Value(now),
			); err != nil {
				return err
			}
		}
		if len(retired) > 0 {
			paths := make([]string, 0, len(retired))
			for _, p := range retired {
				paths = append(paths, p.RoutePath)
			}
			if _, err := qp.WithContext(ctx).Where(qp.RoutePath.In(paths...)).UpdateSimple(
				qp.Status.Value("inactive"),
				qp.UpdatedAt.Value(now),
			); err != nil {
				return err
			}
		}
		if err := s.auditSvc.RecordSuccess(ctx, tx, actor, "sync_pages", "pages", "sync", "sync", nil, resp); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return vo.PageSyncVO{}, err
	}
	return resp, nil
}

func normalizePageSource(s string, def string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	switch s {
	case "router", "menu", "manual":
		return s
	default:
		return def
	}
}
