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

var _ dtoresp.ServiceListResponse
var _ dtoresp.ServiceGetResponse

type ServiceController struct {
	svc *service.ServiceService
}

func NewServiceController(svc *service.ServiceService) *ServiceController {
	return &ServiceController{svc: svc}
}

func (ctl *ServiceController) Register(rg *gin.RouterGroup) {
	rg.GET("/hosts/:id/services", ctl.listByHost)
	rg.GET("/services", ctl.list)
	rg.GET("/services/:id", ctl.get)
	rg.POST("/services", ctl.create)
	rg.PUT("/services/:id", ctl.update)
	rg.DELETE("/services/:id", ctl.delete)
}

// @Summary 列出指定主机下的服务
// @Tags hosts
// @Security BearerAuth
// @Produce json
// @Param id path int true "Host ID"
// @Param page query int false "Page (>=1); invalid -> 400"
// @Param page_size query int false "Page size (1~100); invalid -> 400"
// @Param keyword query string false "Search keyword (name like)"
// @Param status query string false "Status; invalid -> 400" Enums(active,inactive)
// @Param service_kind query string false "Service kind; invalid -> 400" Enums(app,api,database,middleware,cloud_product,other)
// @Param environment_id query int false "Environment ID; invalid -> 400"
// @Success 200 {object} dtoresp.ServiceListResponse
// @Failure 400 {object} ErrorResponse
// @Router /hosts/{id}/services [get]
func (ctl *ServiceController) listByHost(c *gin.Context) {
	hostID, ok := parseUint64Param(c, "id")
	if !ok {
		return
	}
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
	serviceKind, ok := parseQueryEnumStrict(c, "service_kind", map[string]struct{}{
		"app":           {},
		"api":           {},
		"database":      {},
		"middleware":    {},
		"cloud_product": {},
		"other":         {},
	})
	if !ok {
		return
	}
	filter := service.ServiceListFilter{
		Keyword:     strings.TrimSpace(c.Query("keyword")),
		Status:      status,
		ServiceKind: serviceKind,
		HostID:      &hostID,
	}
	if id, ok := parseQueryUint64PtrStrict(c, "environment_id"); !ok {
		return
	} else {
		filter.EnvironmentID = id
	}

	list, total, err := ctl.svc.ListPaged(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.ServiceListData{
		List: list,
		Pagination: dtoresp.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	})
}

// @Summary 获取服务详情
// @Tags services
// @Security BearerAuth
// @Produce json
// @Param id path int true "Service ID"
// @Success 200 {object} dtoresp.ServiceGetResponse
// @Router /services/{id} [get]
func (ctl *ServiceController) get(c *gin.Context) {
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

// @Summary 列出服务
// @Tags services
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page (>=1); invalid -> 400"
// @Param page_size query int false "Page size (1~100); invalid -> 400"
// @Param keyword query string false "Search keyword (name like)"
// @Param status query string false "Status; invalid -> 400" Enums(active,inactive)
// @Param service_kind query string false "Service kind; invalid -> 400" Enums(app,api,database,middleware,cloud_product,other)
// @Param host_id query int false "Host ID; invalid -> 400"
// @Param environment_id query int false "Environment ID; invalid -> 400"
// @Success 200 {object} dtoresp.ServiceListResponse
// @Failure 400 {object} ErrorResponse
// @Router /services [get]
func (ctl *ServiceController) list(c *gin.Context) {
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
	serviceKind, ok := parseQueryEnumStrict(c, "service_kind", map[string]struct{}{
		"app":           {},
		"api":           {},
		"database":      {},
		"middleware":    {},
		"cloud_product": {},
		"other":         {},
	})
	if !ok {
		return
	}
	filter := service.ServiceListFilter{
		Keyword:     strings.TrimSpace(c.Query("keyword")),
		Status:      status,
		ServiceKind: serviceKind,
	}
	if id, ok := parseQueryUint64PtrStrict(c, "host_id"); !ok {
		return
	} else {
		filter.HostID = id
	}
	if id, ok := parseQueryUint64PtrStrict(c, "environment_id"); !ok {
		return
	} else {
		filter.EnvironmentID = id
	}

	list, total, err := ctl.svc.ListPaged(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.ServiceListData{
		List: list,
		Pagination: dtoresp.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	})
}

// @Summary 创建服务
// @Tags services
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body request.ServiceCreateRequest true "ServiceCreateRequest (environment_ids required, [] allowed)"
// @Success 200 {object} dtoresp.IDResponse
// @Router /services [post]
func (ctl *ServiceController) create(c *gin.Context) {
	var req request.ServiceCreateRequest
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

// @Summary 更新服务
// @Tags services
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Service ID"
// @Param body body request.ServiceUpdateRequest true "ServiceUpdateRequest (environment_ids required, [] allowed)"
// @Success 200 {object} dtoresp.EmptyResponse
// @Router /services/{id} [put]
func (ctl *ServiceController) update(c *gin.Context) {
	id, ok := parseUint64Param(c, "id")
	if !ok {
		return
	}
	var req request.ServiceUpdateRequest
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

// @Summary 删除服务
// @Tags services
// @Security BearerAuth
// @Produce json
// @Param id path int true "Service ID"
// @Success 200 {object} dtoresp.EmptyResponse
// @Router /services/{id} [delete]
func (ctl *ServiceController) delete(c *gin.Context) {
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
