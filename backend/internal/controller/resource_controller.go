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

var _ dtoresp.ResourceListResponse

type ResourceController struct {
	svc *service.ResourceService
}

// NewResourceController 保留用于 compatibility storage。
// Deprecated: 请勿在 resources 三表上扩展新业务能力。
func NewResourceController(svc *service.ResourceService) *ResourceController {
	return &ResourceController{svc: svc}
}

func (ctl *ResourceController) Register(rg *gin.RouterGroup) {
	rg.GET("/resources", ctl.list)
	rg.POST("/resources", ctl.create)
	rg.PUT("/resources/:key", ctl.update)
	rg.DELETE("/resources/:key", ctl.delete)
}

// @Summary [Deprecated][Compatibility] 列出资源
// @Tags resources
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page (>=1); invalid -> 400"
// @Param page_size query int false "Page size (1~100); invalid -> 400"
// @Param keyword query string false "Search keyword (name like)"
// @Param type query string false "Resource type; invalid -> 400" Enums(host,service)
// @Param status query string false "Status; invalid -> 400" Enums(active,inactive)
// @Success 200 {object} dtoresp.ResourceListResponse
// @Failure 400 {object} ErrorResponse
// @Router /resources [get]
func (ctl *ResourceController) list(c *gin.Context) {
	page, pageSize, ok := parsePaginationStrict(c)
	if !ok {
		return
	}
	tp, ok := parseQueryEnumStrict(c, "type", map[string]struct{}{
		"host":    {},
		"service": {},
	})
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
	filter := service.ResourceListFilter{
		Keyword: strings.TrimSpace(c.Query("keyword")),
		Type:    tp,
		Status:  status,
	}
	list, total, err := ctl.svc.ListPaged(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.ResourceListData{
		List: list,
		Pagination: dtoresp.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	})
}

// @Summary [Deprecated][Compatibility] 创建资源
// @Tags resources
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body request.ResourceCreateRequest true "ResourceCreateRequest"
// @Success 200 {object} dtoresp.KeyResponse
// @Router /resources [post]
func (ctl *ResourceController) create(c *gin.Context) {
	var req request.ResourceCreateRequest
	if !bindJSONStrict(c, &req) {
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	key, err := ctl.svc.Create(c.Request.Context(), actor, req)
	if err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.KeyData{Key: key})
}

// @Summary [Deprecated][Compatibility] 更新资源
// @Tags resources
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param key path int true "Resource Key"
// @Param body body request.ResourceUpdateRequest true "ResourceUpdateRequest"
// @Success 200 {object} dtoresp.EmptyResponse
// @Router /resources/{key} [put]
func (ctl *ResourceController) update(c *gin.Context) {
	key, ok := parseUint64Param(c, "key")
	if !ok {
		return
	}
	var req request.ResourceUpdateRequest
	if !bindJSONStrict(c, &req) {
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	if err := ctl.svc.Update(c.Request.Context(), actor, key, req); err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.EmptyData{})
}

// @Summary [Deprecated][Compatibility] 删除资源
// @Tags resources
// @Security BearerAuth
// @Produce json
// @Param key path int true "Resource Key"
// @Success 200 {object} dtoresp.EmptyResponse
// @Router /resources/{key} [delete]
func (ctl *ResourceController) delete(c *gin.Context) {
	key, ok := parseUint64Param(c, "key")
	if !ok {
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	if err := ctl.svc.Delete(c.Request.Context(), actor, key); err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.EmptyData{})
}
