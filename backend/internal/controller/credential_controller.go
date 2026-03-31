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

var _ dtoresp.CredentialListResponse
var _ dtoresp.CredentialRevealResponse

type CredentialController struct {
	svc *service.CredentialService
}

func NewCredentialController(svc *service.CredentialService) *CredentialController {
	return &CredentialController{svc: svc}
}

func (ctl *CredentialController) Register(rg *gin.RouterGroup) {
	rg.GET("/host-credentials", ctl.listHost)
	rg.POST("/host-credentials", ctl.createHost)
	rg.PUT("/host-credentials/:id", ctl.updateHost)
	rg.DELETE("/host-credentials/:id", ctl.deleteHost)
	rg.POST("/host-credentials/:id/reveal", ctl.revealHost)

	rg.GET("/service-credentials", ctl.listService)
	rg.POST("/service-credentials", ctl.createService)
	rg.PUT("/service-credentials/:id", ctl.updateService)
	rg.DELETE("/service-credentials/:id", ctl.deleteService)
	rg.POST("/service-credentials/:id/reveal", ctl.revealService)
}

// @Summary 列出主机凭据
// @Tags credentials
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page (>=1); invalid -> 400"
// @Param page_size query int false "Page size (1~100); invalid -> 400"
// @Param host_id query int true "Host ID; invalid -> 400"
// @Success 200 {object} dtoresp.CredentialListResponse
// @Failure 400 {object} ErrorResponse
// @Router /host-credentials [get]
func (ctl *CredentialController) listHost(c *gin.Context) {
	page, pageSize, ok := parsePaginationStrict(c)
	if !ok {
		return
	}
	hostIDStr := strings.TrimSpace(c.Query("host_id"))
	if hostIDStr == "" {
		failInvalidQuery(c, "host_id", "required")
		return
	}
	hostID, err := parseUint64(hostIDStr)
	if err != nil || hostID == 0 {
		failInvalidQuery(c, "host_id", "invalid")
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	list, total, err := ctl.svc.ListHostCredentials(c.Request.Context(), actor, hostID, page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.CredentialListData{
		List: list,
		Pagination: dtoresp.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	})
}

// @Summary 创建主机凭据
// @Tags credentials
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body request.HostCredentialCreateRequest true "HostCredentialCreateRequest"
// @Success 200 {object} dtoresp.IDResponse
// @Router /host-credentials [post]
func (ctl *CredentialController) createHost(c *gin.Context) {
	var req request.HostCredentialCreateRequest
	if !bindJSONStrict(c, &req) {
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	id, err := ctl.svc.CreateHostCredential(c.Request.Context(), actor, req)
	if err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.IDData{ID: id})
}

// @Summary 更新主机凭据
// @Tags credentials
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Credential ID"
// @Param body body request.HostCredentialUpdateRequest true "HostCredentialUpdateRequest"
// @Success 200 {object} dtoresp.EmptyResponse
// @Router /host-credentials/{id} [put]
func (ctl *CredentialController) updateHost(c *gin.Context) {
	id, ok := parseUint64Param(c, "id")
	if !ok {
		return
	}
	var req request.HostCredentialUpdateRequest
	if !bindJSONStrict(c, &req) {
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	if err := ctl.svc.UpdateHostCredential(c.Request.Context(), actor, id, req); err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.EmptyData{})
}

// @Summary 删除主机凭据
// @Tags credentials
// @Security BearerAuth
// @Produce json
// @Param id path int true "Credential ID"
// @Success 200 {object} dtoresp.EmptyResponse
// @Router /host-credentials/{id} [delete]
func (ctl *CredentialController) deleteHost(c *gin.Context) {
	id, ok := parseUint64Param(c, "id")
	if !ok {
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	if err := ctl.svc.DeleteHostCredential(c.Request.Context(), actor, id); err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.EmptyData{})
}

// @Summary 查看主机凭据明文
// @Tags credentials
// @Security BearerAuth
// @Produce json
// @Param id path int true "Credential ID"
// @Success 200 {object} dtoresp.CredentialRevealResponse
// @Router /host-credentials/{id}/reveal [post]
func (ctl *CredentialController) revealHost(c *gin.Context) {
	id, ok := parseUint64Param(c, "id")
	if !ok {
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	data, err := ctl.svc.RevealHostCredential(c.Request.Context(), actor, id)
	if err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, data)
}

// @Summary 列出服务凭据
// @Tags credentials
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page (>=1); invalid -> 400"
// @Param page_size query int false "Page size (1~100); invalid -> 400"
// @Param service_id query int true "Service ID; invalid -> 400"
// @Success 200 {object} dtoresp.CredentialListResponse
// @Failure 400 {object} ErrorResponse
// @Router /service-credentials [get]
func (ctl *CredentialController) listService(c *gin.Context) {
	page, pageSize, ok := parsePaginationStrict(c)
	if !ok {
		return
	}
	serviceIDStr := strings.TrimSpace(c.Query("service_id"))
	if serviceIDStr == "" {
		failInvalidQuery(c, "service_id", "required")
		return
	}
	serviceID, err := parseUint64(serviceIDStr)
	if err != nil || serviceID == 0 {
		failInvalidQuery(c, "service_id", "invalid")
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	list, total, err := ctl.svc.ListServiceCredentials(c.Request.Context(), actor, serviceID, page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.CredentialListData{
		List: list,
		Pagination: dtoresp.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	})
}

// @Summary 创建服务凭据
// @Tags credentials
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body request.ServiceCredentialCreateRequest true "ServiceCredentialCreateRequest"
// @Success 200 {object} dtoresp.IDResponse
// @Router /service-credentials [post]
func (ctl *CredentialController) createService(c *gin.Context) {
	var req request.ServiceCredentialCreateRequest
	if !bindJSONStrict(c, &req) {
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	id, err := ctl.svc.CreateServiceCredential(c.Request.Context(), actor, req)
	if err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.IDData{ID: id})
}

// @Summary 更新服务凭据
// @Tags credentials
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Credential ID"
// @Param body body request.ServiceCredentialUpdateRequest true "ServiceCredentialUpdateRequest"
// @Success 200 {object} dtoresp.EmptyResponse
// @Router /service-credentials/{id} [put]
func (ctl *CredentialController) updateService(c *gin.Context) {
	id, ok := parseUint64Param(c, "id")
	if !ok {
		return
	}
	var req request.ServiceCredentialUpdateRequest
	if !bindJSONStrict(c, &req) {
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	if err := ctl.svc.UpdateServiceCredential(c.Request.Context(), actor, id, req); err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.EmptyData{})
}

// @Summary 删除服务凭据
// @Tags credentials
// @Security BearerAuth
// @Produce json
// @Param id path int true "Credential ID"
// @Success 200 {object} dtoresp.EmptyResponse
// @Router /service-credentials/{id} [delete]
func (ctl *CredentialController) deleteService(c *gin.Context) {
	id, ok := parseUint64Param(c, "id")
	if !ok {
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	if err := ctl.svc.DeleteServiceCredential(c.Request.Context(), actor, id); err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.EmptyData{})
}

// @Summary 查看服务凭据明文
// @Tags credentials
// @Security BearerAuth
// @Produce json
// @Param id path int true "Credential ID"
// @Success 200 {object} dtoresp.CredentialRevealResponse
// @Router /service-credentials/{id}/reveal [post]
func (ctl *CredentialController) revealService(c *gin.Context) {
	id, ok := parseUint64Param(c, "id")
	if !ok {
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	data, err := ctl.svc.RevealServiceCredential(c.Request.Context(), actor, id)
	if err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, data)
}
