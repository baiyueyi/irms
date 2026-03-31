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

var _ dtoresp.HostListResponse
var _ dtoresp.HostGetResponse

type HostController struct {
	svc *service.HostService
}

func NewHostController(svc *service.HostService) *HostController {
	return &HostController{svc: svc}
}

func (ctl *HostController) Register(rg *gin.RouterGroup) {
	rg.GET("/hosts", ctl.list)
	rg.GET("/hosts/:id", ctl.get)
	rg.POST("/hosts", ctl.create)
	rg.PUT("/hosts/:id", ctl.update)
	rg.DELETE("/hosts/:id", ctl.delete)
}

// @Summary 获取主机详情
// @Tags hosts
// @Security BearerAuth
// @Produce json
// @Param id path int true "Host ID"
// @Success 200 {object} dtoresp.HostGetResponse
// @Router /hosts/{id} [get]
func (ctl *HostController) get(c *gin.Context) {
	id, ok := parseUint64Param(c, "id")
	if !ok {
		return
	}
	item, err := ctl.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, item)
}

// @Summary 列出主机
// @Tags hosts
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page (>=1); invalid -> 400"
// @Param page_size query int false "Page size (1~100); invalid -> 400"
// @Param keyword query string false "Search keyword (name/hostname like)"
// @Param status query string false "Status; invalid -> 400" Enums(active,inactive)
// @Param provider_kind query string false "Provider kind; invalid -> 400" Enums(physical,vm,cloud_instance,other)
// @Param location_id query int false "Location ID; invalid -> 400"
// @Param environment_id query int false "Environment ID; invalid -> 400"
// @Success 200 {object} dtoresp.HostListResponse
// @Failure 400 {object} ErrorResponse
// @Router /hosts [get]
func (ctl *HostController) list(c *gin.Context) {
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
	providerKind, ok := parseQueryEnumStrict(c, "provider_kind", map[string]struct{}{
		"physical":       {},
		"vm":             {},
		"cloud_instance": {},
		"other":          {},
	})
	if !ok {
		return
	}
	filter := service.HostListFilter{
		Keyword:      strings.TrimSpace(c.Query("keyword")),
		Status:       status,
		ProviderKind: providerKind,
	}
	if id, ok := parseQueryUint64PtrStrict(c, "location_id"); !ok {
		return
	} else {
		filter.LocationID = id
	}
	if id, ok := parseQueryUint64PtrStrict(c, "environment_id"); !ok {
		return
	} else {
		filter.EnvironmentID = id
	}

	list, total, err := ctl.svc.ListPaged(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		handleError(c, err)
	}
	apiresp.OK(c, dtoresp.HostListData{
		List: list,
		Pagination: dtoresp.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	})
}

// @Summary 创建主机
// @Tags hosts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body request.HostCreateRequest true "HostCreateRequest (environment_ids required, [] allowed)"
// @Success 200 {object} dtoresp.IDResponse
// @Router /hosts [post]
func (ctl *HostController) create(c *gin.Context) {
	var req request.HostCreateRequest
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

// @Summary 更新主机
// @Tags hosts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Host ID"
// @Param body body request.HostUpdateRequest true "HostUpdateRequest (environment_ids required, [] allowed)"
// @Success 200 {object} dtoresp.EmptyResponse
// @Router /hosts/{id} [put]
func (ctl *HostController) update(c *gin.Context) {
	id, ok := parseUint64Param(c, "id")
	if !ok {
		return
	}
	var req request.HostUpdateRequest
	if !bindJSONStrict(c, &req) {
		return
	}
	if strings.TrimSpace(req.Status) == "" {
		req.Status = "active"
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

// @Summary 删除主机
// @Tags hosts
// @Security BearerAuth
// @Produce json
// @Param id path int true "Host ID"
// @Success 200 {object} dtoresp.EmptyResponse
// @Router /hosts/{id} [delete]
func (ctl *HostController) delete(c *gin.Context) {
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
