package controller

import (
	"net/http"
	"strings"

	"irms/backend/internal/dto/request"
	dtoresp "irms/backend/internal/dto/response"
	"irms/backend/internal/middleware"
	ecode "irms/backend/internal/pkg/errors"
	apiresp "irms/backend/internal/pkg/response"
	"irms/backend/internal/service"

	"github.com/gin-gonic/gin"
)

var _ dtoresp.UserListResponse

type UserController struct {
	svc *service.UserService
}

func NewUserController(svc *service.UserService) *UserController {
	return &UserController{svc: svc}
}

func (ctl *UserController) Register(rg *gin.RouterGroup) {
	rg.GET("/users", ctl.list)
	rg.POST("/users", ctl.create)
	rg.PUT("/users/:id", ctl.update)
	rg.DELETE("/users/:id", ctl.delete)
}

// @Summary 列出用户
// @Tags users
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page (>=1); invalid -> 400"
// @Param page_size query int false "Page size (1~100); invalid -> 400"
// @Param keyword query string false "Search keyword (username like)"
// @Param role query string false "Role; invalid -> 400" Enums(super_admin,user)
// @Param status query string false "Status; invalid -> 400" Enums(enabled,disabled)
// @Success 200 {object} dtoresp.UserListResponse
// @Failure 400 {object} ErrorResponse
// @Router /users [get]
func (ctl *UserController) list(c *gin.Context) {
	page, pageSize, ok := parsePaginationStrict(c)
	if !ok {
		return
	}
	role, ok := parseQueryEnumStrict(c, "role", map[string]struct{}{
		"super_admin": {},
		"user":        {},
	})
	if !ok {
		return
	}
	status, ok := parseQueryEnumStrict(c, "status", map[string]struct{}{
		"enabled":  {},
		"disabled": {},
	})
	if !ok {
		return
	}
	filter := service.UserListFilter{
		Keyword: strings.TrimSpace(c.Query("keyword")),
		Status:  status,
		Role:    role,
	}
	list, total, err := ctl.svc.ListPaged(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.UserListData{
		List: list,
		Pagination: dtoresp.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	})
}

// @Summary 创建用户
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body request.UserCreateRequest true "UserCreateRequest"
// @Success 200 {object} dtoresp.IDResponse
// @Router /users [post]
func (ctl *UserController) create(c *gin.Context) {
	var req request.UserCreateRequest
	if !bindJSONStrict(c, &req) {
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	id, err := ctl.svc.Create(c.Request.Context(), actor, req.Username, req.Password, req.Role, req.Status)
	if err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.IDData{ID: id})
}

// @Summary 更新用户
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param body body request.UserUpdateRequest true "UserUpdateRequest"
// @Success 200 {object} dtoresp.EmptyResponse
// @Router /users/{id} [put]
func (ctl *UserController) update(c *gin.Context) {
	id, ok := parseUint64Param(c, "id")
	if !ok {
		return
	}
	var req request.UserUpdateRequest
	if !bindJSONStrict(c, &req) {
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	if err := ctl.svc.Update(c.Request.Context(), actor, id, req.Role, req.Status); err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.EmptyData{})
}

// @Summary 删除用户
// @Tags users
// @Security BearerAuth
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} dtoresp.EmptyResponse
// @Router /users/{id} [delete]
func (ctl *UserController) delete(c *gin.Context) {
	id, ok := parseUint64Param(c, "id")
	if !ok {
		return
	}
	cu, ok := middleware.CurrentUserFromContext(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	if id == cu.ID {
		apiresp.Fail(c, http.StatusBadRequest, ecode.CodeInvalidArgument, "cannot delete self", nil)
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	if err := ctl.svc.Delete(c.Request.Context(), actor, id); err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.EmptyData{})
}
