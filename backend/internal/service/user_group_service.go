package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	"irms/backend/internal/model"
	"irms/backend/internal/query"
	"irms/backend/internal/store"
	"irms/backend/internal/vo"

	"gorm.io/gorm"
)

type UserGroupService struct {
	db       *gorm.DB
	q        *query.Query
	auditSvc *AuditService
}

func NewUserGroupService(st *store.Store) *UserGroupService {
	return &UserGroupService{
		db:       st.Gorm,
		q:        st.Query,
		auditSvc: NewAuditService(st.Gorm),
	}
}

func (s *UserGroupService) ListPaged(ctx context.Context, keyword string, page int, pageSize int) ([]vo.UserGroupVO, int, error) {
	qg := s.q.UserGroup
	do := qg.WithContext(ctx)
	if keyword != "" {
		do = do.Where(qg.Name.Like("%" + keyword + "%"))
	}
	items, total64, err := do.Order(qg.ID.Desc()).FindByPage((page-1)*pageSize, pageSize)
	if err != nil {
		return nil, 0, err
	}
	ids := make([]uint64, 0, len(items))
	for _, it := range items {
		ids = append(ids, it.ID)
	}
	countByID := map[uint64]int{}
	if len(ids) > 0 {
		qugm := s.q.UserGroupMember
		members, err := qugm.WithContext(ctx).Select(qugm.UserGroupID).Where(qugm.UserGroupID.In(ids...)).Find()
		if err != nil {
			return nil, 0, err
		}
		for _, m := range members {
			countByID[m.UserGroupID]++
		}
	}
	out := make([]vo.UserGroupVO, 0, len(items))
	for _, it := range items {
		desc := ""
		if it.Description != nil {
			desc = *it.Description
		}
		out = append(out, vo.UserGroupVO{
			ID:          it.ID,
			Name:        it.Name,
			Description: desc,
			MemberCount: countByID[it.ID],
			CreatedAt:   it.CreatedAt,
			UpdatedAt:   it.UpdatedAt,
		})
	}
	return out, int(total64), nil
}

func (s *UserGroupService) Create(ctx context.Context, actor Actor, name string, description string) (uint64, error) {
	now := time.Now()
	item := model.UserGroup{
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if description != "" {
		v := description
		item.Description = &v
	}
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		if err := qtx.UserGroup.WithContext(ctx).Create(&item); err != nil {
			return err
		}
		after := struct {
			Name string `json:"name"`
		}{Name: name}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "create_user_group", "user_groups", strconv.FormatUint(item.ID, 10), name, nil, after)
	}); err != nil {
		return 0, err
	}
	return item.ID, nil
}

func (s *UserGroupService) Update(ctx context.Context, actor Actor, id uint64, name string, description string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		qug := qtx.UserGroup
		before, err := qug.WithContext(ctx).Where(qug.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if _, err := qug.WithContext(ctx).Where(qug.ID.Eq(id)).UpdateSimple(
			qug.Name.Value(name),
			qug.UpdatedAt.Value(time.Now()),
		); err != nil {
			return err
		}
		if _, err := qug.WithContext(ctx).Where(qug.ID.Eq(id)).Update(qug.Description, nullOrString(description)); err != nil {
			return err
		}
		after := struct {
			Name string `json:"name"`
		}{Name: name}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "update_user_group", "user_groups", strconv.FormatUint(id, 10), name, before, after)
	})
}

func (s *UserGroupService) Delete(ctx context.Context, actor Actor, id uint64) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		qug := qtx.UserGroup
		before, err := qug.WithContext(ctx).Where(qug.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if _, err := qug.WithContext(ctx).Where(qug.ID.Eq(id)).Delete(); err != nil {
			return err
		}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "delete_user_group", "user_groups", strconv.FormatUint(id, 10), before.Name, before, nil)
	})
}

func (s *UserGroupService) ListMembers(ctx context.Context, groupID uint64, keyword string, page int, pageSize int) ([]vo.UserGroupMemberVO, int, error) {
	qugm := s.q.UserGroupMember
	do := qugm.WithContext(ctx).Where(qugm.UserGroupID.Eq(groupID))
	if v := strings.TrimSpace(keyword); v != "" {
		qu := s.q.User
		users, err := qu.WithContext(ctx).Select(qu.ID).Where(qu.Username.Like("%" + v + "%")).Find()
		if err != nil {
			return nil, 0, err
		}
		if len(users) == 0 {
			return []vo.UserGroupMemberVO{}, 0, nil
		}
		userIDs := make([]uint64, 0, len(users))
		for _, u := range users {
			userIDs = append(userIDs, u.ID)
		}
		do = do.Where(qugm.UserID.In(userIDs...))
	}
	items, total64, err := do.Order(qugm.ID.Desc()).FindByPage((page-1)*pageSize, pageSize)
	if err != nil {
		return nil, 0, err
	}
	userIDsSet := map[uint64]struct{}{}
	for _, it := range items {
		userIDsSet[it.UserID] = struct{}{}
	}
	userIDs := setToSlice(userIDsSet)
	qu := s.q.User
	users, err := qu.WithContext(ctx).Select(qu.ID, qu.Username).Where(qu.ID.In(userIDs...)).Find()
	if err != nil {
		return nil, 0, err
	}
	usernameByID := map[uint64]string{}
	for _, u := range users {
		usernameByID[u.ID] = u.Username
	}
	out := make([]vo.UserGroupMemberVO, 0, len(items))
	for _, it := range items {
		out = append(out, vo.UserGroupMemberVO{
			ID:          it.ID,
			UserID:      it.UserID,
			UserGroupID: it.UserGroupID,
			UserName:    usernameByID[it.UserID],
			CreatedAt:   it.CreatedAt,
		})
	}
	return out, int(total64), nil
}

func (s *UserGroupService) AddMember(ctx context.Context, actor Actor, userID uint64, groupID uint64) error {
	now := time.Now()
	qugm := s.q.UserGroupMember
	return qugm.WithContext(ctx).Create(&model.UserGroupMember{UserID: userID, UserGroupID: groupID, CreatedAt: now})
}

func (s *UserGroupService) RemoveMember(ctx context.Context, actor Actor, userID uint64, groupID uint64) error {
	qugm := s.q.UserGroupMember
	if _, err := qugm.WithContext(ctx).Where(qugm.UserID.Eq(userID), qugm.UserGroupID.Eq(groupID)).Delete(); err != nil {
		return err
	}
	return nil
}
