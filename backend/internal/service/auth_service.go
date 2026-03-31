package service

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"irms/backend/internal/config"
	ecode "irms/backend/internal/pkg/errors"
	"irms/backend/internal/query"
	"irms/backend/internal/vo"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	cfg      config.Config
	db       *gorm.DB
	auditSvc *AuditService
}

func NewAuthService(cfg config.Config, db *gorm.DB) *AuthService {
	return &AuthService{
		cfg:      cfg,
		db:       db,
		auditSvc: NewAuditService(db),
	}
}

func (s *AuthService) Login(ctx context.Context, username string, password string) (token string, user vo.UserVO, err error) {
	q := query.Use(s.db)
	qu := q.User
	u, err := qu.WithContext(ctx).Where(qu.Username.Eq(username)).First()
	if err != nil {
		return "", vo.UserVO{}, err
	}
	if u.Status != "enabled" {
		return "", vo.UserVO{}, errors.New("user disabled")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", vo.UserVO{}, errors.New("invalid credentials")
	}
	claims := jwt.MapClaims{
		"user_id":  u.ID,
		"username": u.Username,
		"role":     u.Role,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := t.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", vo.UserVO{}, err
	}
	return tokenStr, toUserVO(*u), nil
}

func (s *AuthService) ChangePassword(ctx context.Context, actor Actor, userID uint64, oldPassword string, newPassword string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qtx := query.Use(tx)
		qu := qtx.User

		u, err := qu.WithContext(ctx).Where(qu.ID.Eq(userID)).First()
		if err != nil {
			_ = s.auditSvc.RecordFailure(ctx, nil, actor, "change_password", "users", "0", "0", nil, struct {
				Error string `json:"error"`
			}{Error: "user_not_found"})
			return err
		}
		if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(oldPassword)); err != nil {
			_ = s.auditSvc.RecordFailure(ctx, nil, actor, "change_password", "users", strconv.FormatUint(userID, 10), u.Username, struct {
				ID               uint64 `json:"id"`
				Username         string `json:"username"`
				MustChangePasswd bool   `json:"must_change_password"`
			}{ID: u.ID, Username: u.Username, MustChangePasswd: u.MustChangePassword}, struct {
				Error string `json:"error"`
			}{Error: "old_password_mismatch"})
			return ecode.NewAppError(http.StatusUnauthorized, ecode.CodeInvalidCredentials, "old password mismatch", nil)
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			_ = s.auditSvc.RecordFailure(ctx, nil, actor, "change_password", "users", strconv.FormatUint(userID, 10), u.Username, nil, struct {
				Error string `json:"error"`
			}{Error: "hash_failed"})
			return err
		}
		if _, err := qu.WithContext(ctx).Where(qu.ID.Eq(userID)).UpdateSimple(
			qu.PasswordHash.Value(string(hash)),
			qu.MustChangePassword.Value(false),
			qu.UpdatedAt.Value(time.Now()),
		); err != nil {
			return err
		}
		return s.auditSvc.RecordSuccess(ctx, tx, actor, "change_password", "users", strconv.FormatUint(userID, 10), u.Username, struct {
			ID               uint64 `json:"id"`
			Username         string `json:"username"`
			MustChangePasswd bool   `json:"must_change_password"`
		}{ID: u.ID, Username: u.Username, MustChangePasswd: u.MustChangePassword}, struct {
			MustChangePasswd bool `json:"must_change_password"`
		}{MustChangePasswd: false})
	})
}

func (s *AuthService) Me(ctx context.Context, userID uint64) (vo.UserVO, error) {
	q := query.Use(s.db)
	qu := q.User
	u, err := qu.WithContext(ctx).Where(qu.ID.Eq(userID)).First()
	if err != nil {
		return vo.UserVO{}, err
	}
	return toUserVO(*u), nil
}
