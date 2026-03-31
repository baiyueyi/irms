package controller

import (
	"net/http"
	"strings"

	"irms/backend/internal/dto/request"
	dtoresp "irms/backend/internal/dto/response"
	ecode "irms/backend/internal/pkg/errors"
	apiresp "irms/backend/internal/pkg/response"
	"irms/backend/internal/service"

	"github.com/gin-gonic/gin"
)

var _ dtoresp.EnvironmentListResponse

type EnvironmentController struct {
	svc *service.EnvironmentService
}

func NewEnvironmentController(svc *service.EnvironmentService) *EnvironmentController {
	return &EnvironmentController{svc: svc}
}

func (ctl *EnvironmentController) Register(rg *gin.RouterGroup) {
	rg.GET("/environments", ctl.list)
	rg.POST("/environments", ctl.create)
	rg.PUT("/environments/:id", ctl.update)
	rg.DELETE("/environments/:id", ctl.delete)
}

// @Summary 列出环境
// @Tags environments
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page (>=1); invalid -> 400"
// @Param page_size query int false "Page size (1~100); invalid -> 400"
// @Param keyword query string false "Search keyword (name like)"
// @Success 200 {object} dtoresp.EnvironmentListResponse
// @Failure 400 {object} ErrorResponse
// @Router /environments [get]
func (ctl *EnvironmentController) list(c *gin.Context) {
	page, pageSize, ok := parsePaginationStrict(c)
	if !ok {
		return
	}
	keyword := strings.TrimSpace(c.Query("keyword"))
	list, total, err := ctl.svc.ListPaged(c.Request.Context(), keyword, page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.EnvironmentListData{
		List: list,
		Pagination: dtoresp.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	})
}

// @Summary 创建环境
// @Tags environments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body request.EnvironmentCreateRequest true "EnvironmentCreateRequest"
// @Success 200 {object} dtoresp.IDResponse
// @Router /environments [post]
func (ctl *EnvironmentController) create(c *gin.Context) {
	var req request.EnvironmentCreateRequest
	if !bindJSONStrict(c, &req) {
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	id, err := ctl.svc.Create(c.Request.Context(), actor, req)
	if err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.IDData{ID: id})
}

// @Summary 更新环境
// @Tags environments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Environment ID"
// @Param body body request.EnvironmentUpdateRequest true "EnvironmentUpdateRequest"
// @Success 200 {object} dtoresp.EmptyResponse
// @Router /environments/{id} [put]
func (ctl *EnvironmentController) update(c *gin.Context) {
	id, ok := parseUint64Param(c, "id")
	if !ok {
		return
	}
	var req request.EnvironmentUpdateRequest
	if !bindJSONStrict(c, &req) {
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	if err := ctl.svc.Update(c.Request.Context(), actor, id, req); err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.EmptyData{})
}

// @Summary 删除环境
// @Tags environments
// @Security BearerAuth
// @Produce json
// @Param id path int true "Environment ID"
// @Success 200 {object} dtoresp.EmptyResponse
// @Router /environments/{id} [delete]
func (ctl *EnvironmentController) delete(c *gin.Context) {
	id, ok := parseUint64Param(c, "id")
	if !ok {
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
