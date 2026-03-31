package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"irms/backend/internal/config"
	"irms/backend/internal/query"
	ecode "irms/backend/internal/pkg/errors"
	"irms/backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type CurrentUser struct {
	ID               uint64
	Username         string
	Role             string
	MustChangePasswd bool
}

func AuthRequired(cfg config.Config, q *query.Query) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "missing authorization", nil)
			c.Abort()
			return
		}
		if !strings.HasPrefix(authHeader, "Bearer ") {
			response.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "invalid authorization", nil)
			c.Abort()
			return
		}
		tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(cfg.JWTSecret), nil
		})
		if err != nil || token == nil || !token.Valid {
			response.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "invalid token", nil)
			c.Abort()
			return
		}
		rawID, ok := claims["user_id"]
		if !ok {
			response.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "invalid token", nil)
			c.Abort()
			return
		}
		userID, ok := toUint64(rawID)
		if !ok || userID == 0 {
			response.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "invalid token", nil)
			c.Abort()
			return
		}
		qu := q.User
		u, err := qu.WithContext(c.Request.Context()).Where(qu.ID.Eq(userID)).First()
		if err != nil {
			response.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "invalid user", nil)
			c.Abort()
			return
		}
		if u.Status != "enabled" {
			response.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "user disabled", nil)
			c.Abort()
			return
		}
		path := c.Request.URL.Path
		if u.MustChangePassword && path != "/api/auth/change-password" && path != "/api/me" {
			response.Fail(c, http.StatusForbidden, ecode.CodeFirstLoginPasswordChange, "first login password change required", nil)
			c.Abort()
			return
		}
		c.Set("current_user", CurrentUser{
			ID:               u.ID,
			Username:         u.Username,
			Role:             u.Role,
			MustChangePasswd: u.MustChangePassword,
		})
		c.Next()
	}
}

func SuperAdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		u, ok := CurrentUserFromContext(c)
		if !ok || u.Role != "super_admin" {
			response.Fail(c, http.StatusForbidden, ecode.CodeForbidden, "forbidden", nil)
			c.Abort()
			return
		}
		c.Next()
	}
}

func CurrentUserFromContext(c *gin.Context) (CurrentUser, bool) {
	v, ok := c.Get("current_user")
	if !ok {
		return CurrentUser{}, false
	}
	u, ok := v.(CurrentUser)
	return u, ok
}

func toUint64(v interface{}) (uint64, bool) {
	switch t := v.(type) {
	case float64:
		if t < 0 {
			return 0, false
		}
		return uint64(t), true
	case int64:
		if t < 0 {
			return 0, false
		}
		return uint64(t), true
	case uint64:
		return t, true
	case json.Number:
		i, err := t.Int64()
		if err != nil || i < 0 {
			return 0, false
		}
		return uint64(i), true
	default:
		return 0, false
	}
}
