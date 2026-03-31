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

type ResourceGroupService struct {
	st       *store.Store
	db       *gorm.DB
	q        *query.Query
	auditSvc *AuditService
}

func NewResourceGroupService(st *store.Store) *ResourceGroupService {
	return &ResourceGroupService{
		st:       st,
		db:       st.Gorm,
		q:        st.Query,
		auditSvc: NewAuditService(st.Gorm),
	}
}

type ResourceGroupListFilter struct {
	Keyword string
	Type    string
}

func (s *ResourceGroupService) ListPaged(ctx context.Context, f ResourceGroupListFilter, page int, pageSize int) ([]vo.ResourceGroupVO, int, error) {
	qrg := s.q.ResourceGroup
	do := qrg.WithContext(ctx)
	if v := strings.TrimSpace(f.Keyword); v != "" {
		do = do.Where(qrg.Name.Like("%" + v + "%"))
	}
	if v := strings.TrimSpace(f.Type); v != "" {
		do = do.Where(qrg.Type.Eq(v))
	}
	rows, total64, err := do.Order(qrg.ID.Desc()).FindByPage((page-1)*pageSize, pageSize)
	if err != nil {
		return nil, 0, err
	}
	memberCountByGroupID := map[uint64]int{}
	if len(rows) > 0 {
		groupIDs := make([]uint64, 0, len(rows))
		for _, r := range rows {
			groupIDs = append(groupIDs, r.ID)
		}
		qrgm := s.q.ResourceGroupMember
		members, err := qrgm.WithContext(ctx).
			Select(qrgm.ResourceGroupID).
			Where(qrgm.ResourceGroupID.In(groupIDs...)).
			Find()
		if err != nil {
			return nil, 0, err
		}
		for _, m := range members {
			memberCountByGroupID[m.ResourceGroupID]++
		}
	}
	out := make([]vo.ResourceGroupVO, 0, len(rows))
	for _, r := range rows {
		desc := ""
		if r.Description != nil {
			desc = *r.Description
		}
		out = append(out, vo.ResourceGroupVO{
			ID:          r.ID,
			Name:        r.Name,
			GroupType:   toExternalGroupType(r.Type),
			Description: desc,
			MemberCount: memberCountByGroupID[r.ID],
			CreatedAt:   r.CreatedAt,
			UpdatedAt:   r.UpdatedAt,
		})
	}
	return out, int(total64), nil
}

func (s *ResourceGroupService) Create(ctx context.Context, actor Actor, req request.ResourceGroupCreateRequest) (uint64, error) {
	if !validResourceGroupType(req.Type) {
		return 0, ecode.NewAppError(http.StatusBadRequest, ecode.CodeInvalidArgument, "invalid group_type, must be host_group or service_group", nil)
	}
	now := time.Now()
	it := model.ResourceGroup{
		Name:      strings.TrimSpace(req.Name),
		Type:      strings.TrimSpace(req.Type),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if strings.TrimSpace(req.Description) != "" {
		v := strings.TrimSpace(req.Description)
		it.Description = &v
	}
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		if err := qtx.ResourceGroup.WithContext(ctx).Create(&it); err != nil {
			if dbutil.IsDuplicate(err) {
				return ecode.NewAppError(http.StatusConflict, ecode.CodeConflict, "group name exists", nil)
			}
			return err
		}
		targetType := toExternalGroupType(it.Type)
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "create_"+targetType, targetType, strconv.FormatUint(it.ID, 10), it.Name, nil, req)
	}); err != nil {
		return 0, err
	}
	return it.ID, nil
}

func (s *ResourceGroupService) Update(ctx context.Context, actor Actor, id uint64, req request.ResourceGroupUpdateRequest) error {
	if !validResourceGroupType(req.Type) {
		return ecode.NewAppError(http.StatusBadRequest, ecode.CodeInvalidArgument, "invalid group_type, must be host_group or service_group", nil)
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		stx := s.st.WithTx(tx)
		qrg := stx.Query.ResourceGroup
		before, err := qrg.WithContext(ctx).Where(qrg.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if _, err := qrg.WithContext(ctx).Where(qrg.ID.Eq(id)).UpdateSimple(
			qrg.Name.Value(strings.TrimSpace(req.Name)),
			qrg.Type.Value(strings.TrimSpace(req.Type)),
			qrg.UpdatedAt.Value(time.Now()),
		); err != nil {
			if dbutil.IsDuplicate(err) {
				return ecode.NewAppError(http.StatusConflict, ecode.CodeConflict, "group name exists", nil)
			}
			return err
		}
		if _, err := qrg.WithContext(ctx).Where(qrg.ID.Eq(id)).Update(qrg.Description, nullOrString(strings.TrimSpace(req.Description))); err != nil {
			return err
		}
		targetType := toExternalGroupType(req.Type)
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "update_"+targetType, targetType, strconv.FormatUint(id, 10), req.Name, before, req)
	})
}

func (s *ResourceGroupService) Delete(ctx context.Context, actor Actor, id uint64) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		stx := s.st.WithTx(tx)
		qrg := stx.Query.ResourceGroup
		before, err := qrg.WithContext(ctx).Where(qrg.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if _, err := qrg.WithContext(ctx).Where(qrg.ID.Eq(id)).Delete(); err != nil {
			return err
		}
		targetType := toExternalGroupType(before.Type)
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "delete_"+targetType, targetType, strconv.FormatUint(id, 10), before.Name, before, nil)
	})
}

func (s *ResourceGroupService) ListMembersPaged(ctx context.Context, groupID uint64, groupType string, keyword string, page int, pageSize int) ([]vo.ResourceGroupMemberVO, int, error) {
	qrg := s.q.ResourceGroup
	groupItem, err := qrg.WithContext(ctx).Where(qrg.ID.Eq(groupID)).First()
	if err != nil {
		return []vo.ResourceGroupMemberVO{}, 0, nil
	}
	if groupType != "" && groupItem.Type != groupType {
		return []vo.ResourceGroupMemberVO{}, 0, nil
	}
	qrgm := s.q.ResourceGroupMember
	do := qrgm.WithContext(ctx).Where(qrgm.ResourceGroupID.Eq(groupID))
	if v := strings.TrimSpace(keyword); v != "" {
		qr := s.q.Resource
		relatedResources, err := qr.WithContext(ctx).Select(qr.Key).Where(qr.Name.Like("%" + v + "%")).Find()
		if err != nil {
			return nil, 0, err
		}
		if len(relatedResources) == 0 {
			return []vo.ResourceGroupMemberVO{}, 0, nil
		}
		keys := make([]uint64, 0, len(relatedResources))
		for _, it := range relatedResources {
			keys = append(keys, it.Key)
		}
		do = do.Where(qrgm.ResourceKey.In(keys...))
	}
	rows, total64, err := do.Order(qrgm.ID.Desc()).FindByPage((page-1)*pageSize, pageSize)
	if err != nil {
		return nil, 0, err
	}
	resourceNameByKey := map[uint64]string{}
	resourceTypeByKey := map[uint64]string{}
	if len(rows) > 0 {
		keys := make([]uint64, 0, len(rows))
		for _, r := range rows {
			keys = append(keys, r.ResourceKey)
		}
		qr := s.q.Resource
		resources, err := qr.WithContext(ctx).Select(qr.Key, qr.Name, qr.Type).Where(qr.Key.In(keys...)).Find()
		if err != nil {
			return nil, 0, err
		}
		for _, r := range resources {
			resourceNameByKey[r.Key] = r.Name
			resourceTypeByKey[r.Key] = r.Type
		}
	}
	out := make([]vo.ResourceGroupMemberVO, 0, len(rows))
	groupTypeVO := toExternalGroupType(groupItem.Type)
	for _, r := range rows {
		out = append(out, vo.ResourceGroupMemberVO{
			ID:         r.ID,
			GroupID:    r.ResourceGroupID,
			GroupType:  groupTypeVO,
			MemberID:   r.ResourceKey,
			MemberType: resourceTypeByKey[r.ResourceKey],
			MemberName: resourceNameByKey[r.ResourceKey],
			CreatedAt:  r.CreatedAt,
		})
	}
	return out, int(total64), nil
}

func (s *ResourceGroupService) AddMember(ctx context.Context, actor Actor, memberID uint64, groupID uint64, memberType string) error {
	normalizedMemberType, ok := normalizeMemberTypeInput(memberType)
	if !ok {
		return ecode.NewAppError(http.StatusBadRequest, ecode.CodeInvalidArgument, "invalid member_type, must be host or service", nil)
	}
	qr := s.q.Resource
	resourceItem, err := qr.WithContext(ctx).Where(qr.Key.Eq(memberID)).First()
	if err != nil {
		return err
	}
	qrg := s.q.ResourceGroup
	groupItem, err := qrg.WithContext(ctx).Where(qrg.ID.Eq(groupID)).First()
	if err != nil {
		return ecode.NewAppError(http.StatusNotFound, ecode.CodeNotFound, "resource or group not found", nil)
	}
	if normalizedMemberType != "" && resourceItem.Type != normalizedMemberType {
		return ecode.NewAppError(http.StatusBadRequest, ecode.CodeInvalidArgument, "member_type does not match member_id", nil)
	}
	if resourceItem.Type != groupItem.Type {
		return ecode.NewAppError(http.StatusBadRequest, ecode.CodeResourceTypeMismatch, "member_type does not match group_type", nil)
	}
	it := model.ResourceGroupMember{
		ResourceKey:     memberID,
		ResourceGroupID: groupID,
		CreatedAt:       time.Now(),
	}
	if err := s.q.ResourceGroupMember.WithContext(ctx).Create(&it); err != nil {
		if dbutil.IsDuplicate(err) {
			return ecode.NewAppError(http.StatusConflict, ecode.CodeConflict, "member exists", nil)
		}
		return err
	}
	targetType := toExternalGroupType(groupItem.Type)
	_ = s.auditSvc.RecordSuccess(ctx, nil, actor, "add_"+targetType+"_member", targetType, strconv.FormatUint(groupID, 10), groupItem.Name, nil, struct {
		GroupID    uint64 `json:"group_id"`
		GroupType  string `json:"group_type"`
		MemberID   uint64 `json:"member_id"`
		MemberType string `json:"member_type"`
		MemberName string `json:"member_name"`
	}{
		GroupID:    groupID,
		GroupType:  targetType,
		MemberID:   memberID,
		MemberType: resourceItem.Type,
		MemberName: resourceItem.Name,
	})
	return nil
}

func (s *ResourceGroupService) RemoveMember(ctx context.Context, actor Actor, memberID uint64, groupID uint64, memberType string) error {
	normalizedMemberType, ok := normalizeMemberTypeInput(memberType)
	if !ok {
		return ecode.NewAppError(http.StatusBadRequest, ecode.CodeInvalidArgument, "invalid member_type, must be host or service", nil)
	}
	qr := s.q.Resource
	resourceItem, err := qr.WithContext(ctx).Where(qr.Key.Eq(memberID)).First()
	if err != nil {
		return err
	}
	if normalizedMemberType != "" && resourceItem.Type != normalizedMemberType {
		return ecode.NewAppError(http.StatusBadRequest, ecode.CodeInvalidArgument, "member_type does not match member_id", nil)
	}
	qrg := s.q.ResourceGroup
	groupItem, err := qrg.WithContext(ctx).Where(qrg.ID.Eq(groupID)).First()
	if err != nil {
		return ecode.NewAppError(http.StatusNotFound, ecode.CodeNotFound, "group not found", nil)
	}
	if resourceItem.Type != groupItem.Type {
		return ecode.NewAppError(http.StatusBadRequest, ecode.CodeResourceTypeMismatch, "member_type does not match group_type", nil)
	}
	qrgm := s.q.ResourceGroupMember
	_, err = qrgm.WithContext(ctx).
		Where(qrgm.ResourceKey.Eq(memberID), qrgm.ResourceGroupID.Eq(groupID)).
		Delete()
	if err != nil {
		return err
	}
	targetType := toExternalGroupType(groupItem.Type)
	_ = s.auditSvc.RecordSuccess(ctx, nil, actor, "remove_"+targetType+"_member", targetType, strconv.FormatUint(groupID, 10), groupItem.Name, struct {
		GroupID    uint64 `json:"group_id"`
		GroupType  string `json:"group_type"`
		MemberID   uint64 `json:"member_id"`
		MemberType string `json:"member_type"`
		MemberName string `json:"member_name"`
	}{
		GroupID:    groupID,
		GroupType:  targetType,
		MemberID:   memberID,
		MemberType: resourceItem.Type,
		MemberName: resourceItem.Name,
	}, nil)
	return nil
}

func validResourceGroupType(t string) bool {
	return t == "host" || t == "service"
}

func toExternalGroupType(groupType string) string {
	switch strings.TrimSpace(groupType) {
	case "host":
		return "host_group"
	case "service":
		return "service_group"
	default:
		return strings.TrimSpace(groupType)
	}
}

func normalizeMemberTypeInput(memberType string) (string, bool) {
	switch strings.TrimSpace(strings.ToLower(memberType)) {
	case "":
		return "", true
	case "host", "host_group":
		return "host", true
	case "service", "service_group":
		return "service", true
	default:
		return "", false
	}
}
