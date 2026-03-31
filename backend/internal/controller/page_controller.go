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

var _ dtoresp.PageListResponse
var _ dtoresp.PageSyncResponse

type PageController struct {
	svc *service.PageService
}

func NewPageController(svc *service.PageService) *PageController {
	return &PageController{svc: svc}
}

func (ctl *PageController) Register(rg *gin.RouterGroup) {
	rg.GET("/pages", ctl.list)
	rg.POST("/pages", ctl.create)
	rg.PUT("/pages/:id", ctl.update)
	rg.DELETE("/pages/:id", ctl.delete)
	rg.POST("/pages/sync", ctl.sync)
}

// @Summary 列出页面资源
// @Tags pages
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page (>=1); invalid -> 400"
// @Param page_size query int false "Page size (1~100); invalid -> 400"
// @Param keyword query string false "Search keyword (name like)"
// @Param status query string false "Status; invalid -> 400" Enums(active,inactive)
// @Success 200 {object} dtoresp.PageListResponse
// @Failure 400 {object} ErrorResponse
// @Router /pages [get]
func (ctl *PageController) list(c *gin.Context) {
	page, pageSize, ok := parsePaginationStrict(c)
	if !ok {
		return
	}
	status, ok := parseQueryEnumStrict(c, "status", map[string]struct{}{
		"active":   {},
		"inactive": {},
	})
	if !ok {
		return
	}
	keyword := strings.TrimSpace(c.Query("keyword"))
	list, total, err := ctl.svc.ListPaged(c.Request.Context(), keyword, status, page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.PageListData{
		List: list,
		Pagination: dtoresp.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	})
}

// @Summary 创建页面资源
// @Tags pages
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body request.PageCreateRequest true "PageCreateRequest"
// @Success 200 {object} dtoresp.IDResponse
// @Router /pages [post]
func (ctl *PageController) create(c *gin.Context) {
	var req request.PageCreateRequest
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

// @Summary 更新页面资源
// @Tags pages
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Page ID"
// @Param body body request.PageUpdateRequest true "PageUpdateRequest"
// @Success 200 {object} dtoresp.EmptyResponse
// @Router /pages/{id} [put]
func (ctl *PageController) update(c *gin.Context) {
	id, ok := parseUint64Param(c, "id")
	if !ok {
		return
	}
	var req request.PageUpdateRequest
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

// @Summary 删除页面资源
// @Tags pages
// @Security BearerAuth
// @Produce json
// @Param id path int true "Page ID"
// @Success 200 {object} dtoresp.EmptyResponse
// @Router /pages/{id} [delete]
func (ctl *PageController) delete(c *gin.Context) {
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

// @Summary 同步页面路由
// @Tags pages
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body request.PageSyncRequest true "PageSyncRequest"
// @Success 200 {object} dtoresp.PageSyncResponse
// @Router /pages/sync [post]
func (ctl *PageController) sync(c *gin.Context) {
	var req request.PageSyncRequest
	if !bindJSONStrict(c, &req) {
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	data, err := ctl.svc.Sync(c.Request.Context(), actor, req)
	if err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, data)
}
