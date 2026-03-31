package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	"irms/backend/internal/dto/request"
	"irms/backend/internal/model"
	"irms/backend/internal/query"
	"irms/backend/internal/store"
	"irms/backend/internal/vo"

	"gorm.io/gorm"
)

type HostService struct {
	db        *gorm.DB
	q         *query.Query
	auditSvc  *AuditService
}

func NewHostService(st *store.Store) *HostService {
	return &HostService{
		db:        st.Gorm,
		q:         st.Query,
		auditSvc:  NewAuditService(st.Gorm),
	}
}

func (s *HostService) ListPaged(ctx context.Context, f HostListFilter, page int, pageSize int) ([]vo.HostVO, int, error) {
	qh := s.q.Host
	do := qh.WithContext(ctx)
	if v := strings.TrimSpace(f.Keyword); v != "" {
		do = do.Where(qh.Name.Like("%" + v + "%"))
	}
	if v := strings.TrimSpace(f.Status); v != "" {
		do = do.Where(qh.Status.Eq(v))
	}
	if v := strings.TrimSpace(f.ProviderKind); v != "" {
		do = do.Where(qh.ProviderKind.Eq(v))
	}
	if f.LocationID != nil {
		do = do.Where(qh.LocationID.Eq(*f.LocationID))
	}
	if f.EnvironmentID != nil {
		qhe := s.q.HostEnvironment
		heRows, err := qhe.WithContext(ctx).Select(qhe.HostID).Where(qhe.EnvironmentID.Eq(*f.EnvironmentID)).Find()
		if err != nil {
			return nil, 0, err
		}
		hostIDs := make([]uint64, 0, len(heRows))
		seen := map[uint64]struct{}{}
		for _, r := range heRows {
			if _, ok := seen[r.HostID]; ok {
				continue
			}
			seen[r.HostID] = struct{}{}
			hostIDs = append(hostIDs, r.HostID)
		}
		if len(hostIDs) == 0 {
			return []vo.HostVO{}, 0, nil
		}
		do = do.Where(qh.ID.In(hostIDs...))
	}
	items, total64, err := do.Order(qh.ID.Desc()).FindByPage((page-1)*pageSize, pageSize)
	if err != nil {
		return nil, 0, err
	}

	hostIDs := make([]uint64, 0, len(items))
	locIDsSet := map[uint64]struct{}{}
	for _, h := range items {
		hostIDs = append(hostIDs, h.ID)
		if h.LocationID != nil {
			locIDsSet[*h.LocationID] = struct{}{}
		}
	}
	locIDs := setToSlice(locIDsSet)
	locNameByID := map[uint64]string{}
	if len(locIDs) > 0 {
		ql := s.q.Location
		locItems, err := ql.WithContext(ctx).Select(ql.ID, ql.Name).Where(ql.ID.In(locIDs...)).Find()
		if err != nil {
			return nil, 0, err
		}
		for _, l := range locItems {
			locNameByID[l.ID] = l.Name
		}
	}

	type envRow struct {
		HostID uint64 `gorm:"column:host_id"`
		EnvID  uint64 `gorm:"column:environment_id"`
		Name   string `gorm:"column:name"`
	}
	var envRows []envRow
	if len(hostIDs) > 0 {
		qhe := s.q.HostEnvironment
		qe := s.q.Environment
		if err := qhe.WithContext(ctx).
			Select(qhe.HostID.As("host_id"), qhe.EnvironmentID.As("environment_id"), qe.Name.As("name")).
			Join(qe, qhe.EnvironmentID.EqCol(qe.ID)).
			Where(qhe.HostID.In(hostIDs...)).
			Order(qe.Name).
			Scan(&envRows); err != nil {
			return nil, 0, err
		}
	}
	envNamesByHostID := map[uint64][]string{}
	envIDsByHostID := map[uint64][]uint64{}
	for _, r := range envRows {
		envNamesByHostID[r.HostID] = append(envNamesByHostID[r.HostID], r.Name)
		envIDsByHostID[r.HostID] = append(envIDsByHostID[r.HostID], r.EnvID)
	}

	out := make([]vo.HostVO, 0, len(items))
	for _, h := range items {
		location := ""
		if h.LocationID != nil {
			location = locNameByID[*h.LocationID]
		}
		desc := ""
		if h.Description != nil {
			desc = *h.Description
		}
		osType := ""
		if h.OsType != nil {
			osType = *h.OsType
		}
		cloudVendor := ""
		if h.CloudVendor != nil {
			cloudVendor = *h.CloudVendor
		}
		cloudInstanceID := ""
		if h.CloudInstanceID != nil {
			cloudInstanceID = *h.CloudInstanceID
		}
		envs := envNamesByHostID[h.ID]
		if envs == nil {
			envs = []string{}
		}
		envIDs := envIDsByHostID[h.ID]
		if envIDs == nil {
			envIDs = []uint64{}
		}
		source := EnvironmentSourceHost
		if len(envIDs) == 0 {
			source = EnvironmentSourceNone
		}
		out = append(out, vo.HostVO{
			ID:              h.ID,
			Name:            h.Name,
			Hostname:        h.Hostname,
			PrimaryAddress:  h.PrimaryAddress,
			ProviderKind:    h.ProviderKind,
			CloudVendor:     cloudVendor,
			CloudInstanceID: cloudInstanceID,
			OSType:          osType,
			Status:          h.Status,
			LocationID:      h.LocationID,
			Location:        location,
			OwnEnvironmentIDs:       envIDs,
			InheritedEnvironmentIDs: []uint64{},
			EnvironmentIDs:  envIDs,
			Environments:    envs,
			SelfEnvironments:       envs,
			EnvironmentSource:      string(source),
			Description:     desc,
			CreatedAt:       h.CreatedAt,
			UpdatedAt:       h.UpdatedAt,
		})
	}
	return out, int(total64), nil
}

func (s *HostService) GetByID(ctx context.Context, id uint64) (vo.HostVO, error) {
	qh := s.q.Host
	h, err := qh.WithContext(ctx).Where(qh.ID.Eq(id)).First()
	if err != nil {
		return vo.HostVO{}, err
	}
	location := ""
	if h.LocationID != nil {
		ql := s.q.Location
		if loc, err := ql.WithContext(ctx).Where(ql.ID.Eq(*h.LocationID)).First(); err == nil {
			location = loc.Name
		}
	}

	type envRow struct {
		EnvID uint64 `gorm:"column:environment_id"`
		Name  string `gorm:"column:name"`
	}
	var envRows []envRow
	qhe := s.q.HostEnvironment
	qe := s.q.Environment
	if err := qhe.WithContext(ctx).
		Select(qhe.EnvironmentID.As("environment_id"), qe.Name.As("name")).
		Join(qe, qhe.EnvironmentID.EqCol(qe.ID)).
		Where(qhe.HostID.Eq(id)).
		Order(qe.Name).
		Scan(&envRows); err != nil {
		return vo.HostVO{}, err
	}
	envNames := make([]string, 0, len(envRows))
	envIDs := make([]uint64, 0, len(envRows))
	for _, r := range envRows {
		envNames = append(envNames, r.Name)
		envIDs = append(envIDs, r.EnvID)
	}

	desc := ""
	if h.Description != nil {
		desc = *h.Description
	}
	cloudVendor := ""
	if h.CloudVendor != nil {
		cloudVendor = *h.CloudVendor
	}
	cloudInstanceID := ""
	if h.CloudInstanceID != nil {
		cloudInstanceID = *h.CloudInstanceID
	}
	osType := ""
	if h.OsType != nil {
		osType = *h.OsType
	}

	source := EnvironmentSourceHost
	if len(envIDs) == 0 {
		source = EnvironmentSourceNone
	}
	return vo.HostVO{
		ID:                    h.ID,
		Name:                  h.Name,
		Hostname:              h.Hostname,
		PrimaryAddress:        h.PrimaryAddress,
		ProviderKind:          h.ProviderKind,
		CloudVendor:           cloudVendor,
		CloudInstanceID:       cloudInstanceID,
		OSType:                osType,
		Status:                h.Status,
		LocationID:            h.LocationID,
		Location:              location,
		OwnEnvironmentIDs:     envIDs,
		InheritedEnvironmentIDs: []uint64{},
		EnvironmentIDs:        envIDs,
		Environments:          envNames,
		SelfEnvironments:      envNames,
		EnvironmentSource:     string(source),
		Description:           desc,
		CreatedAt:             h.CreatedAt,
		UpdatedAt:             h.UpdatedAt,
	}, nil
}

func (s *HostService) Create(ctx context.Context, actor Actor, req request.HostCreateRequest) (uint64, error) {
	now := time.Now()
	status := req.Status
	if status == "" {
		status = "active"
	}
	envIDs := []uint64{}
	if req.EnvironmentIDs != nil {
		envIDs = normalizeUint64List(*req.EnvironmentIDs)
	}
	h := model.Host{
		Name:            req.Name,
		Hostname:        req.Hostname,
		PrimaryAddress:  req.PrimaryAddress,
		ProviderKind:    req.ProviderKind,
		Status:          status,
		LocationID:      req.LocationID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if req.CloudVendor != "" {
		v := req.CloudVendor
		h.CloudVendor = &v
	}
	if req.CloudInstanceID != "" {
		v := req.CloudInstanceID
		h.CloudInstanceID = &v
	}
	if req.OSType != "" {
		v := req.OSType
		h.OsType = &v
	}
	if req.Description != "" {
		v := req.Description
		h.Description = &v
	}
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		qh := qtx.Host
		if err := qh.WithContext(ctx).Create(&h); err != nil {
			return err
		}
		if err := s.replaceEnvironmentsTx(ctx, qtx, h.ID, envIDs); err != nil {
			return err
		}
		if err := s.auditSvc.RecordSuccess(ctx, tx, actor, "create_host", "hosts", strconv.FormatUint(h.ID, 10), req.Name, nil, req); err != nil {
			return err
		}
		if err := s.auditSvc.RecordSuccess(
			ctx,
			tx,
			actor,
			"replace_host_environments",
			"host_environments",
			strconv.FormatUint(h.ID, 10),
			strconv.FormatUint(h.ID, 10),
			struct {
				EnvironmentIDs []uint64 `json:"environment_ids"`
			}{EnvironmentIDs: []uint64{}},
			struct {
				EnvironmentIDs []uint64 `json:"environment_ids"`
			}{EnvironmentIDs: envIDs},
		); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return 0, err
	}
	return h.ID, nil
}

func (s *HostService) Update(ctx context.Context, actor Actor, id uint64, req request.HostUpdateRequest) error {
	status := req.Status
	if status == "" {
		status = "active"
	}
	envIDs := []uint64{}
	if req.EnvironmentIDs != nil {
		envIDs = normalizeUint64List(*req.EnvironmentIDs)
	}
	now := time.Now()
	var locationID interface{}
	if req.LocationID != nil {
		locationID = *req.LocationID
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		qh := qtx.Host
		beforeHost, err := qh.WithContext(ctx).Where(qh.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		qhe := qtx.HostEnvironment
		heRows, err := qhe.WithContext(ctx).
			Select(qhe.EnvironmentID).
			Where(qhe.HostID.Eq(id)).
			Order(qhe.EnvironmentID).
			Find()
		if err != nil {
			return err
		}
		beforeEnvIDs := make([]uint64, 0, len(heRows))
		for _, r := range heRows {
			beforeEnvIDs = append(beforeEnvIDs, r.EnvironmentID)
		}

		if _, err := qh.WithContext(ctx).Where(qh.ID.Eq(id)).UpdateSimple(
			qh.Name.Value(req.Name),
			qh.Hostname.Value(req.Hostname),
			qh.PrimaryAddress.Value(req.PrimaryAddress),
			qh.ProviderKind.Value(req.ProviderKind),
			qh.Status.Value(status),
			qh.UpdatedAt.Value(now),
		); err != nil {
			return err
		}
		if _, err := qh.WithContext(ctx).Where(qh.ID.Eq(id)).Update(qh.CloudVendor, nullOrString(req.CloudVendor)); err != nil {
			return err
		}
		if _, err := qh.WithContext(ctx).Where(qh.ID.Eq(id)).Update(qh.CloudInstanceID, nullOrString(req.CloudInstanceID)); err != nil {
			return err
		}
		if _, err := qh.WithContext(ctx).Where(qh.ID.Eq(id)).Update(qh.OsType, nullOrString(req.OSType)); err != nil {
			return err
		}
		if _, err := qh.WithContext(ctx).Where(qh.ID.Eq(id)).Update(qh.LocationID, locationID); err != nil {
			return err
		}
		if _, err := qh.WithContext(ctx).Where(qh.ID.Eq(id)).Update(qh.Description, nullOrString(req.Description)); err != nil {
			return err
		}
		if err := s.replaceEnvironmentsTx(ctx, qtx, id, envIDs); err != nil {
			return err
		}
		before := struct {
			Host          model.Host `json:"host"`
			EnvironmentIDs []uint64  `json:"environment_ids"`
		}{Host: *beforeHost, EnvironmentIDs: beforeEnvIDs}
		after := struct {
			Request request.HostUpdateRequest `json:"request"`
		}{Request: req}
		if err := s.auditSvc.RecordSuccess(ctx, tx, actor, "update_host", "hosts", strconv.FormatUint(id, 10), req.Name, before, after); err != nil {
			return err
		}
		if err := s.auditSvc.RecordSuccess(
			ctx,
			tx,
			actor,
			"replace_host_environments",
			"host_environments",
			strconv.FormatUint(id, 10),
			strconv.FormatUint(id, 10),
			struct {
				EnvironmentIDs []uint64 `json:"environment_ids"`
			}{EnvironmentIDs: beforeEnvIDs},
			struct {
				EnvironmentIDs []uint64 `json:"environment_ids"`
			}{EnvironmentIDs: envIDs},
		); err != nil {
			return err
		}
		return nil
	})
}

func (s *HostService) replaceEnvironmentsTx(ctx context.Context, qtx *query.Query, hostID uint64, envIDs []uint64) error {
	qhe := qtx.HostEnvironment
	if _, err := qhe.WithContext(ctx).Where(qhe.HostID.Eq(hostID)).Delete(); err != nil {
		return err
	}
	if len(envIDs) == 0 {
		return nil
	}
	now := time.Now()
	rows := make([]*model.HostEnvironment, 0, len(envIDs))
	for _, eid := range envIDs {
		rows = append(rows, &model.HostEnvironment{HostID: hostID, EnvironmentID: eid, CreatedAt: now})
	}
	return qhe.WithContext(ctx).Create(rows...)
}

func (s *HostService) Delete(ctx context.Context, actor Actor, id uint64) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		qh := qtx.Host
		beforeHost, err := qh.WithContext(ctx).Where(qh.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if _, err := qh.WithContext(ctx).Where(qh.ID.Eq(id)).Delete(); err != nil {
			return err
		}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "delete_host", "hosts", strconv.FormatUint(id, 10), beforeHost.Name, beforeHost, nil)
	})
}

func normalizeUint64List(in []uint64) []uint64 {
	set := map[uint64]struct{}{}
	out := make([]uint64, 0, len(in))
	for _, v := range in {
		if v == 0 {
			continue
		}
		if _, ok := set[v]; ok {
			continue
		}
		set[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
