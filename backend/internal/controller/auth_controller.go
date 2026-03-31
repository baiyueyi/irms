package controller

import (
	"net/http"

	"irms/backend/internal/dto/request"
	dtoresp "irms/backend/internal/dto/response"
	"irms/backend/internal/middleware"
	ecode "irms/backend/internal/pkg/errors"
	apiresp "irms/backend/internal/pkg/response"
	"irms/backend/internal/service"

	"github.com/gin-gonic/gin"
)

var _ dtoresp.LoginResponse
var _ dtoresp.MeResponse

type AuthController struct {
	svc *service.AuthService
}

func NewAuthController(svc *service.AuthService) *AuthController {
	return &AuthController{svc: svc}
}

func (ctl *AuthController) RegisterPublic(rg *gin.RouterGroup) {
	rg.POST("/auth/login", ctl.login)
}

func (ctl *AuthController) RegisterAuthed(rg *gin.RouterGroup) {
	rg.POST("/auth/change-password", ctl.changePassword)
	rg.GET("/me", ctl.me)
}

// @Summary 登录
// @Tags auth
// @Accept json
// @Produce json
// @Param body body request.LoginRequest true "LoginRequest"
// @Success 200 {object} dtoresp.LoginResponse
// @Router /auth/login [post]
func (ctl *AuthController) login(c *gin.Context) {
	var req request.LoginRequest
	if !bindJSONStrict(c, &req) {
		return
	}
	token, user, err := ctl.svc.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeInvalidCredentials, "invalid username or password", nil)
		return
	}
	apiresp.OK(c, dtoresp.LoginData{Token: token, User: user})
}

// @Summary 修改密码
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body request.ChangePasswordRequest true "ChangePasswordRequest"
// @Success 200 {object} dtoresp.EmptyResponse
// @Router /auth/change-password [post]
func (ctl *AuthController) changePassword(c *gin.Context) {
	var req request.ChangePasswordRequest
	if !bindJSONStrict(c, &req) {
		return
	}
	cu, ok := middleware.CurrentUserFromContext(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	if err := ctl.svc.ChangePassword(c.Request.Context(), actor, cu.ID, req.OldPassword, req.NewPassword); err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.EmptyData{})
}

// @Summary 获取当前用户
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dtoresp.MeResponse
// @Router /me [get]
func (ctl *AuthController) me(c *gin.Context) {
	cu, ok := middleware.CurrentUserFromContext(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	u, err := ctl.svc.Me(c.Request.Context(), cu.ID)
	if err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, u)
}
