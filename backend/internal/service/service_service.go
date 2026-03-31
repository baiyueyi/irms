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

type ServiceService struct {
	db       *gorm.DB
	q        *query.Query
	auditSvc *AuditService
}

func NewServiceService(st *store.Store) *ServiceService {
	return &ServiceService{
		db:       st.Gorm,
		q:        st.Query,
		auditSvc: NewAuditService(st.Gorm),
	}
}

type ServiceListFilter struct {
	Keyword       string
	Status        string
	ServiceKind   string
	HostID        *uint64
	EnvironmentID *uint64
}

func (s *ServiceService) ListPaged(ctx context.Context, f ServiceListFilter, page int, pageSize int) ([]vo.ServiceVO, int, error) {
	qs := s.q.Service
	do := qs.WithContext(ctx)
	if v := strings.TrimSpace(f.Keyword); v != "" {
		do = do.Where(qs.Name.Like("%" + v + "%"))
	}
	if v := strings.TrimSpace(f.Status); v != "" {
		do = do.Where(qs.Status.Eq(v))
	}
	if v := strings.TrimSpace(f.ServiceKind); v != "" {
		do = do.Where(qs.ServiceKind.Eq(v))
	}
	if f.HostID != nil {
		do = do.Where(qs.HostID.Eq(*f.HostID))
	}
	if f.EnvironmentID != nil {
		qse := s.q.ServiceEnvironment
		seRows, err := qse.WithContext(ctx).
			Select(qse.ServiceID).
			Where(qse.EnvironmentID.Eq(*f.EnvironmentID)).
			Find()
		if err != nil {
			return nil, 0, err
		}
		ids := make([]uint64, 0, len(seRows))
		seen := map[uint64]struct{}{}
		for _, r := range seRows {
			if _, ok := seen[r.ServiceID]; ok {
				continue
			}
			seen[r.ServiceID] = struct{}{}
			ids = append(ids, r.ServiceID)
		}
		if len(ids) == 0 {
			return []vo.ServiceVO{}, 0, nil
		}
		do = do.Where(qs.ID.In(ids...))
	}
	items, total64, err := do.Order(qs.ID.Desc()).FindByPage((page-1)*pageSize, pageSize)
	if err != nil {
		return nil, 0, err
	}

	serviceIDs := make([]uint64, 0, len(items))
	hostIDsSet := map[uint64]struct{}{}
	for _, it := range items {
		serviceIDs = append(serviceIDs, it.ID)
		if it.HostID != nil {
			hostIDsSet[*it.HostID] = struct{}{}
		}
	}
	hostIDs := setToSlice(hostIDsSet)
	hostNameByID := map[uint64]string{}
	if len(hostIDs) > 0 {
		qh := s.q.Host
		hostItems, err := qh.WithContext(ctx).Select(qh.ID, qh.Name).Where(qh.ID.In(hostIDs...)).Find()
		if err != nil {
			return nil, 0, err
		}
		for _, h := range hostItems {
			hostNameByID[h.ID] = h.Name
		}
	}

	type envRow struct {
		ServiceID uint64 `gorm:"column:service_id"`
		EnvID     uint64 `gorm:"column:environment_id"`
		Name      string `gorm:"column:name"`
	}
	var svcEnvRows []envRow
	if len(serviceIDs) > 0 {
		qse := s.q.ServiceEnvironment
		qe := s.q.Environment
		if err := qse.WithContext(ctx).
			Select(qse.ServiceID.As("service_id"), qse.EnvironmentID.As("environment_id"), qe.Name.As("name")).
			Join(qe, qse.EnvironmentID.EqCol(qe.ID)).
			Where(qse.ServiceID.In(serviceIDs...)).
			Order(qe.Name).
			Scan(&svcEnvRows); err != nil {
			return nil, 0, err
		}
	}
	ownEnvNamesByServiceID := map[uint64][]string{}
	ownEnvIDsByServiceID := map[uint64][]uint64{}
	for _, r := range svcEnvRows {
		ownEnvNamesByServiceID[r.ServiceID] = append(ownEnvNamesByServiceID[r.ServiceID], r.Name)
		ownEnvIDsByServiceID[r.ServiceID] = append(ownEnvIDsByServiceID[r.ServiceID], r.EnvID)
	}

	type hostEnvRow struct {
		HostID uint64 `gorm:"column:host_id"`
		EnvID  uint64 `gorm:"column:environment_id"`
		Name   string `gorm:"column:name"`
	}
	var hostEnvRows []hostEnvRow
	if len(hostIDs) > 0 {
		qhe := s.q.HostEnvironment
		qe := s.q.Environment
		if err := qhe.WithContext(ctx).
			Select(qhe.HostID.As("host_id"), qhe.EnvironmentID.As("environment_id"), qe.Name.As("name")).
			Join(qe, qhe.EnvironmentID.EqCol(qe.ID)).
			Where(qhe.HostID.In(hostIDs...)).
			Order(qe.Name).
			Scan(&hostEnvRows); err != nil {
			return nil, 0, err
		}
	}
	hostEnvNamesByHostID := map[uint64][]string{}
	hostEnvIDsByHostID := map[uint64][]uint64{}
	for _, r := range hostEnvRows {
		hostEnvNamesByHostID[r.HostID] = append(hostEnvNamesByHostID[r.HostID], r.Name)
		hostEnvIDsByHostID[r.HostID] = append(hostEnvIDsByHostID[r.HostID], r.EnvID)
	}

	out := make([]vo.ServiceVO, 0, len(items))
	for _, it := range items {
		hostName := ""
		inherited := []string{}
		inheritedIDs := []uint64{}
		if it.HostID != nil {
			hostName = hostNameByID[*it.HostID]
			inherited = hostEnvNamesByHostID[*it.HostID]
			inheritedIDs = hostEnvIDsByHostID[*it.HostID]
			if inherited == nil {
				inherited = []string{}
			}
			if inheritedIDs == nil {
				inheritedIDs = []uint64{}
			}
		}
		own := ownEnvNamesByServiceID[it.ID]
		ownIDs := ownEnvIDsByServiceID[it.ID]
		if own == nil {
			own = []string{}
		}
		if ownIDs == nil {
			ownIDs = []uint64{}
		}

		envs := own
		selfEnvs := own
		effectiveIDs := ownIDs
		source := EnvironmentSourceService
		if len(own) == 0 {
			envs = inherited
			selfEnvs = []string{}
			effectiveIDs = inheritedIDs
			if it.HostID != nil {
				if len(inherited) > 0 {
					source = EnvironmentSourceHostInherited
				} else {
					source = EnvironmentSourceNone
				}
			} else {
				source = EnvironmentSourceNone
			}
		}

		desc := ""
		if it.Description != nil {
			desc = *it.Description
		}
		cloudVendor := ""
		if it.CloudVendor != nil {
			cloudVendor = *it.CloudVendor
		}
		cloudProductCode := ""
		if it.CloudProductCode != nil {
			cloudProductCode = *it.CloudProductCode
		}
		protocol := ""
		if it.Protocol != nil {
			protocol = *it.Protocol
		}
		var port *int
		if it.Port != nil {
			p := int(*it.Port)
			port = &p
		}

		out = append(out, vo.ServiceVO{
			ID:                      it.ID,
			Name:                    it.Name,
			ServiceKind:             it.ServiceKind,
			HostID:                  it.HostID,
			Host:                    hostName,
			EndpointOrIdentifier:    it.EndpointOrIdentifier,
			Port:                    port,
			Protocol:                protocol,
			CloudVendor:             cloudVendor,
			CloudProductCode:        cloudProductCode,
			Status:                  it.Status,
			Description:             desc,
			CreatedAt:               it.CreatedAt,
			UpdatedAt:               it.UpdatedAt,
			OwnEnvironmentIDs:       ownIDs,
			InheritedEnvironmentIDs: inheritedIDs,
			EnvironmentIDs:          effectiveIDs,
			Environments:            envs,
			SelfEnvironments:        selfEnvs,
			EnvironmentSource:       string(source),
		})
	}
	return out, int(total64), nil
}

func (s *ServiceService) GetByID(ctx context.Context, id uint64) (vo.ServiceVO, error) {
	qs := s.q.Service
	it, err := qs.WithContext(ctx).Where(qs.ID.Eq(id)).First()
	if err != nil {
		return vo.ServiceVO{}, err
	}

	hostName := ""
	inherited := []string{}
	inheritedIDs := []uint64{}
	if it.HostID != nil {
		qh := s.q.Host
		if h, err := qh.WithContext(ctx).Where(qh.ID.Eq(*it.HostID)).First(); err == nil {
			hostName = h.Name
		}
		type hostEnvRow struct {
			EnvID uint64 `gorm:"column:environment_id"`
			Name  string `gorm:"column:name"`
		}
		var hostEnvRows []hostEnvRow
		qhe := s.q.HostEnvironment
		qe := s.q.Environment
		if err := qhe.WithContext(ctx).
			Select(qhe.EnvironmentID.As("environment_id"), qe.Name.As("name")).
			Join(qe, qhe.EnvironmentID.EqCol(qe.ID)).
			Where(qhe.HostID.Eq(*it.HostID)).
			Order(qe.Name).
			Scan(&hostEnvRows); err != nil {
			return vo.ServiceVO{}, err
		}
		inherited = make([]string, 0, len(hostEnvRows))
		inheritedIDs = make([]uint64, 0, len(hostEnvRows))
		for _, r := range hostEnvRows {
			inherited = append(inherited, r.Name)
			inheritedIDs = append(inheritedIDs, r.EnvID)
		}
	}

	type svcEnvRow struct {
		EnvID uint64 `gorm:"column:environment_id"`
		Name  string `gorm:"column:name"`
	}
	var svcEnvRows []svcEnvRow
	qse := s.q.ServiceEnvironment
	qe := s.q.Environment
	if err := qse.WithContext(ctx).
		Select(qse.EnvironmentID.As("environment_id"), qe.Name.As("name")).
		Join(qe, qse.EnvironmentID.EqCol(qe.ID)).
		Where(qse.ServiceID.Eq(id)).
		Order(qe.Name).
		Scan(&svcEnvRows); err != nil {
		return vo.ServiceVO{}, err
	}

	own := make([]string, 0, len(svcEnvRows))
	ownIDs := make([]uint64, 0, len(svcEnvRows))
	for _, r := range svcEnvRows {
		own = append(own, r.Name)
		ownIDs = append(ownIDs, r.EnvID)
	}

	envs := own
	selfEnvs := own
	effectiveIDs := ownIDs
	source := EnvironmentSourceService
	if len(own) == 0 {
		envs = inherited
		selfEnvs = []string{}
		effectiveIDs = inheritedIDs
		if it.HostID != nil && len(inherited) > 0 {
			source = EnvironmentSourceHostInherited
		} else {
			source = EnvironmentSourceNone
		}
	}

	desc := ""
	if it.Description != nil {
		desc = *it.Description
	}
	cloudVendor := ""
	if it.CloudVendor != nil {
		cloudVendor = *it.CloudVendor
	}
	cloudProductCode := ""
	if it.CloudProductCode != nil {
		cloudProductCode = *it.CloudProductCode
	}
	protocol := ""
	if it.Protocol != nil {
		protocol = *it.Protocol
	}
	var port *int
	if it.Port != nil {
		p := int(*it.Port)
		port = &p
	}

	return vo.ServiceVO{
		ID:                      it.ID,
		Name:                    it.Name,
		ServiceKind:             it.ServiceKind,
		HostID:                  it.HostID,
		Host:                    hostName,
		EndpointOrIdentifier:    it.EndpointOrIdentifier,
		Port:                    port,
		Protocol:                protocol,
		CloudVendor:             cloudVendor,
		CloudProductCode:        cloudProductCode,
		Status:                  it.Status,
		Description:             desc,
		CreatedAt:               it.CreatedAt,
		UpdatedAt:               it.UpdatedAt,
		OwnEnvironmentIDs:       ownIDs,
		InheritedEnvironmentIDs: inheritedIDs,
		EnvironmentIDs:          effectiveIDs,
		Environments:            envs,
		SelfEnvironments:        selfEnvs,
		EnvironmentSource:       string(source),
	}, nil
}
func (s *ServiceService) Create(ctx context.Context, actor Actor, req request.ServiceCreateRequest) (uint64, error) {
	now := time.Now()
	status := req.Status
	if status == "" {
		status = "active"
	}
	envIDs := []uint64{}
	if req.EnvironmentIDs != nil {
		envIDs = normalizeUint64List(*req.EnvironmentIDs)
	}
	var port *int32
	if req.Port != nil {
		p := int32(*req.Port)
		port = &p
	}
	item := model.Service{
		Name:                 req.Name,
		ServiceKind:          req.ServiceKind,
		HostID:               req.HostID,
		EndpointOrIdentifier: req.EndpointOrIdentifier,
		Port:                 port,
		Status:               status,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if req.Protocol != "" {
		v := req.Protocol
		item.Protocol = &v
	}
	if req.CloudVendor != "" {
		v := req.CloudVendor
		item.CloudVendor = &v
	}
	if req.CloudProductCode != "" {
		v := req.CloudProductCode
		item.CloudProductCode = &v
	}
	if req.Description != "" {
		v := req.Description
		item.Description = &v
	}
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		if err := qtx.Service.WithContext(ctx).Create(&item); err != nil {
			return err
		}
		if err := s.replaceEnvironmentsTx(ctx, qtx, item.ID, envIDs); err != nil {
			return err
		}
		if err := s.auditSvc.RecordSuccess(
			ctx,
			tx,
			actor,
			"replace_service_environments",
			"service_environments",
			strconv.FormatUint(item.ID, 10),
			strconv.FormatUint(item.ID, 10),
			struct {
				EnvironmentIDs []uint64 `json:"environment_ids"`
			}{EnvironmentIDs: []uint64{}},
			struct {
				EnvironmentIDs []uint64 `json:"environment_ids"`
			}{EnvironmentIDs: envIDs},
		); err != nil {
			return err
		}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "create_service", "services", strconv.FormatUint(item.ID, 10), req.Name, nil, req)
	}); err != nil {
		return 0, err
	}
	return item.ID, nil
}

func (s *ServiceService) Update(ctx context.Context, actor Actor, id uint64, req request.ServiceUpdateRequest) error {
	envIDs := []uint64{}
	if req.EnvironmentIDs != nil {
		envIDs = normalizeUint64List(*req.EnvironmentIDs)
	}
	now := time.Now()
	var port *int32
	if req.Port != nil {
		p := int32(*req.Port)
		port = &p
	}
	var hostID interface{}
	if req.HostID != nil {
		hostID = *req.HostID
	}
	var portValue interface{}
	if port != nil {
		portValue = *port
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		qs := qtx.Service
		beforeSvc, err := qs.WithContext(ctx).Where(qs.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		qse := qtx.ServiceEnvironment
		type row struct {
			EnvironmentID uint64 `gorm:"column:environment_id"`
		}
		var rows []row
		if err := qse.WithContext(ctx).
			Select(qse.EnvironmentID.As("environment_id")).
			Where(qse.ServiceID.Eq(id)).
			Order(qse.EnvironmentID).
			Scan(&rows); err != nil {
			return err
		}
		beforeEnvIDs := make([]uint64, 0, len(rows))
		for _, r := range rows {
			beforeEnvIDs = append(beforeEnvIDs, r.EnvironmentID)
		}
		if _, err := qs.WithContext(ctx).Where(qs.ID.Eq(id)).UpdateSimple(
			qs.Name.Value(req.Name),
			qs.ServiceKind.Value(req.ServiceKind),
			qs.EndpointOrIdentifier.Value(req.EndpointOrIdentifier),
			qs.Status.Value(req.Status),
			qs.UpdatedAt.Value(now),
		); err != nil {
			return err
		}
		if _, err := qs.WithContext(ctx).Where(qs.ID.Eq(id)).Update(qs.HostID, hostID); err != nil {
			return err
		}
		if _, err := qs.WithContext(ctx).Where(qs.ID.Eq(id)).Update(qs.Port, portValue); err != nil {
			return err
		}
		if _, err := qs.WithContext(ctx).Where(qs.ID.Eq(id)).Update(qs.Protocol, nullOrString(req.Protocol)); err != nil {
			return err
		}
		if _, err := qs.WithContext(ctx).Where(qs.ID.Eq(id)).Update(qs.CloudVendor, nullOrString(req.CloudVendor)); err != nil {
			return err
		}
		if _, err := qs.WithContext(ctx).Where(qs.ID.Eq(id)).Update(qs.CloudProductCode, nullOrString(req.CloudProductCode)); err != nil {
			return err
		}
		if _, err := qs.WithContext(ctx).Where(qs.ID.Eq(id)).Update(qs.Description, nullOrString(req.Description)); err != nil {
			return err
		}
		if err := s.replaceEnvironmentsTx(ctx, qtx, id, envIDs); err != nil {
			return err
		}
		before := struct {
			Service        model.Service `json:"service"`
			EnvironmentIDs []uint64      `json:"environment_ids"`
		}{Service: *beforeSvc, EnvironmentIDs: beforeEnvIDs}
		after := struct {
			Request request.ServiceUpdateRequest `json:"request"`
		}{Request: req}
		if err := s.auditSvc.RecordSuccess(ctx, tx, actor, "update_service", "services", strconv.FormatUint(id, 10), req.Name, before, after); err != nil {
			return err
		}
		if err := s.auditSvc.RecordSuccess(
			ctx,
			tx,
			actor,
			"replace_service_environments",
			"service_environments",
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

func (s *ServiceService) Delete(ctx context.Context, actor Actor, id uint64) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		qs := qtx.Service
		beforeSvc, err := qs.WithContext(ctx).Where(qs.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if _, err := qs.WithContext(ctx).Where(qs.ID.Eq(id)).Delete(); err != nil {
			return err
		}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "delete_service", "services", strconv.FormatUint(id, 10), beforeSvc.Name, beforeSvc, nil)
	})
}

func nullOrString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func (s *ServiceService) replaceEnvironmentsTx(ctx context.Context, qtx *query.Query, serviceID uint64, envIDs []uint64) error {
	qse := qtx.ServiceEnvironment
	if _, err := qse.WithContext(ctx).Where(qse.ServiceID.Eq(serviceID)).Delete(); err != nil {
		return err
	}
	if len(envIDs) == 0 {
		return nil
	}
	now := time.Now()
	rows := make([]*model.ServiceEnvironment, 0, len(envIDs))
	seen := map[uint64]struct{}{}
	for _, id := range envIDs {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		rows = append(rows, &model.ServiceEnvironment{
			ServiceID:     serviceID,
			EnvironmentID: id,
			CreatedAt:     now,
		})
	}
	return qse.WithContext(ctx).Create(rows...)
}
