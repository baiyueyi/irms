package service

import (
	"context"
	"strconv"
	"time"

	"irms/backend/internal/model"
	"irms/backend/internal/query"
	"irms/backend/internal/store"
	"irms/backend/internal/vo"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	db   *gorm.DB
	q    *query.Query
	auditSvc *AuditService
}

type UserListFilter struct {
	Keyword string
	Status  string
	Role    string
}

func NewUserService(st *store.Store) *UserService {
	return &UserService{db: st.Gorm, q: st.Query, auditSvc: NewAuditService(st.Gorm)}
}

func (s *UserService) ListPaged(ctx context.Context, f UserListFilter, page int, pageSize int) ([]vo.UserVO, int, error) {
	qu := s.q.User
	do := qu.WithContext(ctx)
	if f.Keyword != "" {
		do = do.Where(qu.Username.Like("%" + f.Keyword + "%"))
	}
	if f.Role != "" {
		do = do.Where(qu.Role.Eq(f.Role))
	}
	if f.Status != "" {
		do = do.Where(qu.Status.Eq(f.Status))
	}
	items, total64, err := do.Order(qu.ID.Desc()).FindByPage((page-1)*pageSize, pageSize)
	if err != nil {
		return nil, 0, err
	}
	out := make([]vo.UserVO, 0, len(items))
	for _, u := range items {
		out = append(out, toUserVO(*u))
	}
	return out, int(total64), nil
}

func (s *UserService) Create(ctx context.Context, actor Actor, username string, password string, role string, status string) (uint64, error) {
	if role == "" {
		role = "user"
	}
	if status == "" {
		status = "enabled"
	}
	now := time.Now()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	u := model.User{
		Username:         username,
		PasswordHash:     string(hash),
		Role:             role,
		Status:           status,
		MustChangePassword: true,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		qu := qtx.User
		if err := qu.WithContext(ctx).Create(&u); err != nil {
			return err
		}
		after := struct {
			Username string `json:"username"`
			Role     string `json:"role"`
			Status   string `json:"status"`
		}{Username: username, Role: role, Status: status}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "create_user", "users", strconv.FormatUint(u.ID, 10), username, nil, after)
	}); err != nil {
		return 0, err
	}
	return u.ID, nil
}

func (s *UserService) Update(ctx context.Context, actor Actor, id uint64, role string, status string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		qu := qtx.User
		before, err := qu.WithContext(ctx).Where(qu.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if _, err := qu.WithContext(ctx).Where(qu.ID.Eq(id)).UpdateSimple(
			qu.Role.Value(role),
			qu.Status.Value(status),
			qu.UpdatedAt.Value(time.Now()),
		); err != nil {
			return err
		}
		after := struct {
			Role   string `json:"role"`
			Status string `json:"status"`
		}{Role: role, Status: status}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "update_user", "users", strconv.FormatUint(id, 10), before.Username, before, after)
	})
}

func (s *UserService) Delete(ctx context.Context, actor Actor, id uint64) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		qu := qtx.User
		before, err := qu.WithContext(ctx).Where(qu.ID.Eq(id)).First()
		if err != nil {
			return err
		}
		if _, err := qu.WithContext(ctx).Where(qu.ID.Eq(id)).Delete(); err != nil {
			return err
		}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "delete_user", "users", strconv.FormatUint(id, 10), before.Username, before, nil)
	})
}

func toUserVO(u model.User) vo.UserVO {
	return vo.UserVO{
		ID:                 u.ID,
		Username:           u.Username,
		Role:               u.Role,
		Status:             u.Status,
		MustChangePassword: u.MustChangePassword,
		CreatedAt:          u.CreatedAt,
		UpdatedAt:          u.UpdatedAt,
	}
}
