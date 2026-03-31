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

var _ dtoresp.LocationListResponse

type LocationController struct {
	svc *service.LocationService
}

func NewLocationController(svc *service.LocationService) *LocationController {
	return &LocationController{svc: svc}
}

func (ctl *LocationController) Register(rg *gin.RouterGroup) {
	rg.GET("/locations", ctl.list)
	rg.POST("/locations", ctl.create)
	rg.PUT("/locations/:id", ctl.update)
	rg.DELETE("/locations/:id", ctl.delete)
}

// @Summary 列出位置
// @Tags locations
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page (>=1); invalid -> 400"
// @Param page_size query int false "Page size (1~100); invalid -> 400"
// @Param keyword query string false "Search keyword (name like)"
// @Success 200 {object} dtoresp.LocationListResponse
// @Failure 400 {object} ErrorResponse
// @Router /locations [get]
func (ctl *LocationController) list(c *gin.Context) {
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
	apiresp.OK(c, dtoresp.LocationListData{
		List: list,
		Pagination: dtoresp.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	})
}

// @Summary 创建位置
// @Tags locations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body request.LocationCreateRequest true "LocationCreateRequest"
// @Success 200 {object} dtoresp.IDResponse
// @Router /locations [post]
func (ctl *LocationController) create(c *gin.Context) {
	var req request.LocationCreateRequest
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

// @Summary 更新位置
// @Tags locations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Location ID"
// @Param body body request.LocationUpdateRequest true "LocationUpdateRequest"
// @Success 200 {object} dtoresp.EmptyResponse
// @Router /locations/{id} [put]
func (ctl *LocationController) update(c *gin.Context) {
	id, ok := parseUint64Param(c, "id")
	if !ok {
		return
	}
	var req request.LocationUpdateRequest
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

// @Summary 删除位置
// @Tags locations
// @Security BearerAuth
// @Produce json
// @Param id path int true "Location ID"
// @Success 200 {object} dtoresp.EmptyResponse
// @Router /locations/{id} [delete]
func (ctl *LocationController) delete(c *gin.Context) {
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
