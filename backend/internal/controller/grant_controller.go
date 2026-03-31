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

var _ dtoresp.GrantListResponse

type GrantController struct {
	svc *service.GrantService
}

func NewGrantController(svc *service.GrantService) *GrantController {
	return &GrantController{svc: svc}
}

func (ctl *GrantController) Register(rg *gin.RouterGroup) {
	rg.GET("/grants", ctl.list)
	rg.POST("/grants", ctl.upsert)
	rg.PUT("/grants/:id", ctl.update)
	rg.DELETE("/grants/:id", ctl.delete)
}

// @Summary 列出授权关系
// @Tags grants
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page (>=1); invalid -> 400"
// @Param page_size query int false "Page size (1~100); invalid -> 400"
// @Param keyword query string false "Search keyword"
// @Param subject_type query string false "Subject type; invalid -> 400" Enums(user,user_group)
// @Param subject_id query int false "Subject ID; invalid -> 400"
// @Param object_type query string false "Object type; invalid -> 400" Enums(page,host,host_group,service,service_group,host_credential,service_credential)
// @Param object_id query int false "Object ID; invalid -> 400"
// @Param permission query string false "Permission code; invalid -> 400" Enums(page.view,host.read,host.write,service.read,service.write,host_credential.read,host_credential.write,service_credential.read,service_credential.write)
// @Success 200 {object} dtoresp.GrantListResponse
// @Failure 400 {object} ErrorResponse
// @Router /grants [get]
func (ctl *GrantController) list(c *gin.Context) {
	page, pageSize, ok := parsePaginationStrict(c)
	if !ok {
		return
	}
	subjectType, ok := parseQueryEnumStrict(c, "subject_type", map[string]struct{}{
		"user":       {},
		"user_group": {},
	})
	if !ok {
		return
	}
	objectType, ok := parseQueryEnumStrict(c, "object_type", map[string]struct{}{
		"page":               {},
		"host":               {},
		"host_group":         {},
		"service":            {},
		"service_group":      {},
		"host_credential":    {},
		"service_credential": {},
	})
	if !ok {
		return
	}
	permission := strings.TrimSpace(c.Query("permission"))
	filter := service.GrantListFilter{
		Keyword:     strings.TrimSpace(c.Query("keyword")),
		SubjectType: subjectType,
		ObjectType:  objectType,
		Permission:  permission,
	}
	if id, ok := parseQueryUint64PtrStrict(c, "subject_id"); !ok {
		return
	} else {
		filter.SubjectID = id
	}
	if id, ok := parseQueryUint64PtrStrict(c, "object_id"); !ok {
		return
	} else {
		filter.ObjectID = id
	}
	list, total, err := ctl.svc.ListPaged(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.GrantListData{
		List: list,
		Pagination: dtoresp.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	})
}

// @Summary 创建或更新授权关系
// @Tags grants
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body request.GrantUpsertRequest true "GrantUpsertRequest"
// @Success 200 {object} dtoresp.IDResponse
// @Router /grants [post]
func (ctl *GrantController) upsert(c *gin.Context) {
	var req request.GrantUpsertRequest
	if !bindJSONStrict(c, &req) {
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	id, err := ctl.svc.Upsert(c.Request.Context(), actor, req)
	if err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.IDData{ID: id})
}

// @Summary 更新授权关系
// @Tags grants
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Grant ID"
// @Param body body request.GrantUpdateRequest true "GrantUpdateRequest"
// @Success 200 {object} dtoresp.EmptyResponse
// @Router /grants/{id} [put]
func (ctl *GrantController) update(c *gin.Context) {
	id, ok := parseUint64Param(c, "id")
	if !ok {
		return
	}
	var req request.GrantUpdateRequest
	if !bindJSONStrict(c, &req) {
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	if err := ctl.svc.UpdatePermission(c.Request.Context(), actor, id, req.Permission); err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.EmptyData{})
}

// @Summary 删除授权关系
// @Tags grants
// @Security BearerAuth
// @Produce json
// @Param id path int true "Grant ID"
// @Success 200 {object} dtoresp.EmptyResponse
// @Router /grants/{id} [delete]
func (ctl *GrantController) delete(c *gin.Context) {
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
