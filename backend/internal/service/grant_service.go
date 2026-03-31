package service

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"irms/backend/internal/dto/request"
	"irms/backend/internal/pkg/dbutil"
	ecode "irms/backend/internal/pkg/errors"
	"irms/backend/internal/query"
	"irms/backend/internal/repository"
	"irms/backend/internal/store"
	"irms/backend/internal/vo"

	"gorm.io/gorm"
)

type GrantService struct {
	db       *gorm.DB
	repo     *repository.GrantRepository
	q        *query.Query
	permDefs *PermissionDefinitionService
	auditSvc *AuditService
}

func NewGrantService(db *gorm.DB) *GrantService {
	q := query.Use(db)
	return &GrantService{db: db, repo: repository.NewGrantRepository(db), q: q, permDefs: NewPermissionDefinitionService(q), auditSvc: NewAuditService(db)}
}

func NewGrantServiceWithStore(st *store.Store) *GrantService {
	return &GrantService{db: st.Gorm, repo: repository.NewGrantRepository(st.Gorm), q: st.Query, permDefs: NewPermissionDefinitionService(st.Query), auditSvc: NewAuditService(st.Gorm)}
}

type GrantListFilter struct {
	Keyword     string
	SubjectType string
	SubjectID   *uint64
	ObjectType  string
	ObjectID    *uint64
	Permission  string
}

func (s *GrantService) ListPaged(ctx context.Context, f GrantListFilter, page int, pageSize int) ([]vo.GrantVO, int, error) {
	permission := strings.TrimSpace(strings.ToLower(f.Permission))
	if permission != "" {
		ok, err := s.permDefs.ExistsPermissionCode(ctx, permission)
		if err != nil {
			return nil, 0, err
		}
		if !ok {
			return nil, 0, ecode.NewAppError(http.StatusBadRequest, ecode.CodeInvalidArgument, "invalid permission", nil)
		}
	}
	repoFilter := repository.GrantQueryFilter{
		Keyword:     strings.TrimSpace(f.Keyword),
		SubjectType: normalizeGrantSubjectType(strings.TrimSpace(f.SubjectType)),
		SubjectID:   f.SubjectID,
		ObjectType:  normalizeGrantObjectType(strings.TrimSpace(f.ObjectType)),
		ObjectID:    f.ObjectID,
		Permission:  permission,
	}

	total64, err := s.repo.CountByFilter(ctx, repoFilter)
	if err != nil {
		return nil, 0, err
	}
	rows, err := s.repo.ListByFilter(ctx, repoFilter, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	q := s.q

	userIDs := []uint64{}
	userGroupIDs := []uint64{}
	pageIDs := []uint64{}
	hostIDs := []uint64{}
	serviceIDs := []uint64{}
	hostGroupIDs := []uint64{}
	serviceGroupIDs := []uint64{}
	seen := map[string]map[uint64]struct{}{}
	add := func(k string, id uint64, dst *[]uint64) {
		m := seen[k]
		if m == nil {
			m = map[uint64]struct{}{}
			seen[k] = m
		}
		if _, ok := m[id]; ok {
			return
		}
		m[id] = struct{}{}
		*dst = append(*dst, id)
	}
	for _, r := range rows {
		switch r.SubjectType {
		case "user":
			add("u", r.SubjectID, &userIDs)
		case "user_group":
			add("ug", r.SubjectID, &userGroupIDs)
		}
		switch r.ObjectType {
		case "page":
			add("p", r.ObjectID, &pageIDs)
		case "host", "host_credential":
			add("h", r.ObjectID, &hostIDs)
		case "service", "service_credential":
			add("s", r.ObjectID, &serviceIDs)
		case "host_group":
			add("hg", r.ObjectID, &hostGroupIDs)
		case "service_group":
			add("sg", r.ObjectID, &serviceGroupIDs)
		}
	}

	userName := map[uint64]string{}
	if len(userIDs) > 0 {
		u := q.User
		us, err := u.WithContext(ctx).Select(u.ID, u.Username).Where(u.ID.In(userIDs...)).Find()
		if err != nil {
			return nil, 0, err
		}
		for _, u := range us {
			userName[u.ID] = u.Username
		}
	}
	userGroupName := map[uint64]string{}
	if len(userGroupIDs) > 0 {
		ug := q.UserGroup
		ugs, err := ug.WithContext(ctx).Select(ug.ID, ug.Name).Where(ug.ID.In(userGroupIDs...)).Find()
		if err != nil {
			return nil, 0, err
		}
		for _, g := range ugs {
			userGroupName[g.ID] = g.Name
		}
	}
	pageName := map[uint64]string{}
	if len(pageIDs) > 0 {
		p := q.Page
		ps, err := p.WithContext(ctx).Select(p.ID, p.Name).Where(p.ID.In(pageIDs...)).Find()
		if err != nil {
			return nil, 0, err
		}
		for _, p := range ps {
			pageName[p.ID] = p.Name
		}
	}
	hostName := map[uint64]string{}
	if len(hostIDs) > 0 {
		h := q.Host
		hs, err := h.WithContext(ctx).Select(h.ID, h.Name).Where(h.ID.In(hostIDs...)).Find()
		if err != nil {
			return nil, 0, err
		}
		for _, h := range hs {
			hostName[h.ID] = h.Name
		}
	}
	serviceName := map[uint64]string{}
	if len(serviceIDs) > 0 {
		sv := q.Service
		ss, err := sv.WithContext(ctx).Select(sv.ID, sv.Name).Where(sv.ID.In(serviceIDs...)).Find()
		if err != nil {
			return nil, 0, err
		}
		for _, s := range ss {
			serviceName[s.ID] = s.Name
		}
	}
	groupName := map[uint64]string{}
	loadGroupName := func(groupIDs []uint64, groupType string) error {
		if len(groupIDs) == 0 {
			return nil
		}
		rg := q.ResourceGroup
		rgs, err := rg.WithContext(ctx).Select(rg.ID, rg.Name).Where(rg.ID.In(groupIDs...), rg.Type.Eq(groupType)).Find()
		if err != nil {
			return err
		}
		for _, rg := range rgs {
			groupName[rg.ID] = rg.Name
		}
		return nil
	}
	if err := loadGroupName(hostGroupIDs, "host"); err != nil {
		return nil, 0, err
	}
	if err := loadGroupName(serviceGroupIDs, "service"); err != nil {
		return nil, 0, err
	}

	out := make([]vo.GrantVO, 0, len(rows))
	for _, r := range rows {
		stDisp := r.SubjectType
		sName := strconv.FormatUint(r.SubjectID, 10)
		if r.SubjectType == "user" {
			stDisp = "user"
			if v := strings.TrimSpace(userName[r.SubjectID]); v != "" {
				sName = v
			}
		}
		if r.SubjectType == "user_group" {
			stDisp = "group"
			if v := strings.TrimSpace(userGroupName[r.SubjectID]); v != "" {
				sName = v
			}
		}

		otDisp := r.ObjectType
		oName := strconv.FormatUint(r.ObjectID, 10)
		switch r.ObjectType {
		case "page":
			otDisp = "page"
			if v := strings.TrimSpace(pageName[r.ObjectID]); v != "" {
				oName = v
			}
		case "host":
			otDisp = "host"
			if v := strings.TrimSpace(hostName[r.ObjectID]); v != "" {
				oName = v
			}
		case "service":
			otDisp = "service"
			if v := strings.TrimSpace(serviceName[r.ObjectID]); v != "" {
				oName = v
			}
		case "host_credential":
			otDisp = "host_credential"
			if v := strings.TrimSpace(hostName[r.ObjectID]); v != "" {
				oName = v
			}
		case "service_credential":
			otDisp = "service_credential"
			if v := strings.TrimSpace(serviceName[r.ObjectID]); v != "" {
				oName = v
			}
		case "host_group", "service_group":
			if v := strings.TrimSpace(groupName[r.ObjectID]); v != "" {
				oName = v
			}
		}
		out = append(out, vo.GrantVO{
			ID:                 r.ID,
			SubjectType:        r.SubjectType,
			SubjectTypeDisplay: stDisp,
			SubjectID:          r.SubjectID,
			SubjectName:        sName,
			ObjectType:         r.ObjectType,
			ObjectTypeDisplay:  otDisp,
			ObjectID:           r.ObjectID,
			ObjectName:         oName,
			Permission:         r.Permission,
			CreatedAt:          r.CreatedAt,
			UpdatedAt:          r.UpdatedAt,
		})
	}
	return out, int(total64), nil
}

func (s *GrantService) Upsert(ctx context.Context, actor Actor, req request.GrantUpsertRequest) (uint64, error) {
	subjectType := normalizeGrantSubjectType(req.SubjectType)
	objectType := normalizeGrantObjectType(req.ObjectType)
	permission := strings.TrimSpace(strings.ToLower(req.Permission))
	if !isValidGrantSubjectType(subjectType) {
		return 0, ecode.NewAppError(http.StatusBadRequest, ecode.CodeInvalidArgument, "invalid subject type", nil)
	}
	if !isValidGrantObjectType(objectType) {
		return 0, ecode.NewAppError(http.StatusBadRequest, ecode.CodeInvalidArgument, "invalid object type", nil)
	}
	ok, err := s.permDefs.IsValidPermission(ctx, objectType, permission)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, ecode.NewAppError(http.StatusBadRequest, ecode.CodeInvalidArgument, "invalid permission", nil)
	}
	if status, resp, ok := s.validateGrantPrecondition(ctx, subjectType, req.SubjectID, objectType, req.ObjectID, permission); !ok {
		return 0, ecode.NewAppError(status, resp.Code, resp.Message, resp.Details)
	}
	var outID uint64
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		id, err := s.repo.Upsert(ctx, tx, repository.GrantKey{
			SubjectType:    subjectType,
			SubjectID:      req.SubjectID,
			ObjectType:     objectType,
			ObjectID:       req.ObjectID,
			PermissionCode: permission,
		})
		if err != nil {
			if dbutil.IsDuplicate(err) {
				return ecode.NewAppError(http.StatusConflict, ecode.CodeConflict, "grant already exists", nil)
			}
			return err
		}
		outID = id
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "upsert_grant", "grants", strconv.FormatUint(id, 10), strconv.FormatUint(id, 10), nil, req)
	}); err != nil {
		return 0, err
	}
	return outID, nil
}

func (s *GrantService) UpdatePermission(ctx context.Context, actor Actor, id uint64, permission string) error {
	permission = strings.TrimSpace(strings.ToLower(permission))
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		before, err := s.repo.FindByID(ctx, tx, id)
		if err != nil {
			return err
		}
		ok, err := s.permDefs.IsValidPermission(ctx, before.ObjectType, permission)
		if err != nil {
			return err
		}
		if !ok {
			return ecode.NewAppError(http.StatusBadRequest, ecode.CodeInvalidArgument, "invalid permission", nil)
		}
		if status, resp, ok := s.validateGrantPrecondition(ctx, before.SubjectType, before.SubjectID, before.ObjectType, before.ObjectID, permission); !ok {
			return ecode.NewAppError(status, resp.Code, resp.Message, resp.Details)
		}
		if err := s.repo.UpdatePermissionByID(ctx, tx, id, permission); err != nil {
			if dbutil.IsDuplicate(err) {
				return ecode.NewAppError(http.StatusConflict, ecode.CodeConflict, "grant already exists", nil)
			}
			return err
		}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "update_grant", "grants", strconv.FormatUint(id, 10), strconv.FormatUint(id, 10), before, struct {
			Permission string `json:"permission"`
		}{Permission: permission})
	})
}

func (s *GrantService) Delete(ctx context.Context, actor Actor, id uint64) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		qg := qtx.Grant
		before, err := qg.WithContext(ctx).Where(qg.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if _, err := qg.WithContext(ctx).Where(qg.ID.Eq(id)).Delete(); err != nil {
			return err
		}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "revoke_grant", "grants", strconv.FormatUint(id, 10), strconv.FormatUint(id, 10), before, nil)
	})
}

type legacyResp struct {
	Code    string
	Message string
	Details interface{}
}

func (s *GrantService) validateGrantPrecondition(ctx context.Context, subjectType string, subjectID uint64, objectType string, objectID uint64, permission string) (int, legacyResp, bool) {
	switch objectType {
	case "service":
		qs := s.q.Service
		serviceItem, err := qs.WithContext(ctx).Where(qs.ID.Eq(objectID)).First()
		if err != nil {
			return http.StatusInternalServerError, legacyResp{Code: ecode.CodeInternal, Message: "internal error"}, false
		}
		if serviceItem.HostID != nil && *serviceItem.HostID > 0 && (permission == "service.read" || permission == "service.write") {
			ok, err := s.subjectHasAnyPermissionCode(ctx, subjectType, subjectID, "host", *serviceItem.HostID, []string{"host.read", "host.write"})
			if err != nil {
				return http.StatusInternalServerError, legacyResp{Code: ecode.CodeInternal, Message: "internal error"}, false
			}
			if !ok {
				return http.StatusBadRequest, legacyResp{Code: ecode.CodeGrantPreconditionNotMet, Message: "service permission requires host.read precondition"}, false
			}
		}
	case "host_credential":
		qh := s.q.Host
		if _, err := qh.WithContext(ctx).Where(qh.ID.Eq(objectID)).First(); err != nil {
			return http.StatusNotFound, legacyResp{Code: ecode.CodeNotFound, Message: "host not found"}, false
		}
		ok, err := s.subjectHasAnyPermissionCode(ctx, subjectType, subjectID, "host", objectID, []string{"host.read", "host.write"})
		if err != nil {
			return http.StatusInternalServerError, legacyResp{Code: ecode.CodeInternal, Message: "internal error"}, false
		}
		if !ok {
			return http.StatusBadRequest, legacyResp{Code: ecode.CodeGrantPreconditionNotMet, Message: "host_credential permission requires host.read precondition"}, false
		}
	case "service_credential":
		qs := s.q.Service
		if _, err := qs.WithContext(ctx).Where(qs.ID.Eq(objectID)).First(); err != nil {
			return http.StatusNotFound, legacyResp{Code: ecode.CodeNotFound, Message: "service not found"}, false
		}
		ok, err := s.subjectHasAnyPermissionCode(ctx, subjectType, subjectID, "service", objectID, []string{"service.read", "service.write"})
		if err != nil {
			return http.StatusInternalServerError, legacyResp{Code: ecode.CodeInternal, Message: "internal error"}, false
		}
		if !ok {
			return http.StatusBadRequest, legacyResp{Code: ecode.CodeGrantPreconditionNotMet, Message: "service_credential permission requires service.read precondition"}, false
		}
	}
	return http.StatusOK, legacyResp{}, true
}

func normalizeGrantSubjectType(t string) string {
	t = strings.TrimSpace(strings.ToLower(t))
	if t == "group" {
		return "user_group"
	}
	return t
}

func normalizeGrantObjectType(t string) string {
	t = strings.TrimSpace(strings.ToLower(t))
	return t
}

func isValidGrantSubjectType(t string) bool {
	return t == "user" || t == "user_group"
}

func isValidGrantObjectType(t string) bool {
	switch t {
	case "page", "host", "host_group", "service", "service_group", "host_credential", "service_credential":
		return true
	default:
		return false
	}
}

func (s *GrantService) subjectHasAnyPermissionCode(ctx context.Context, subjectType string, subjectID uint64, objectType string, objectID uint64, codes []string) (bool, error) {
	effectiveCodes, err := s.permDefs.ExpandPermissionCodes(ctx, codes)
	if err != nil {
		return false, err
	}
	return s.repo.SubjectHasAnyPermissionCode(ctx, subjectType, subjectID, objectType, objectID, effectiveCodes)
}
