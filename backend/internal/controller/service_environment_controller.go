package controller

import (
	"net/http"

	"irms/backend/internal/dto/request"
	dtoresp "irms/backend/internal/dto/response"
	ecode "irms/backend/internal/pkg/errors"
	apiresp "irms/backend/internal/pkg/response"
	"irms/backend/internal/service"

	"github.com/gin-gonic/gin"
)

var _ dtoresp.ServiceEnvironmentListResponse

type ServiceEnvironmentController struct {
	svc *service.ServiceEnvironmentService
}

func NewServiceEnvironmentController(svc *service.ServiceEnvironmentService) *ServiceEnvironmentController {
	return &ServiceEnvironmentController{svc: svc}
}

func (ctl *ServiceEnvironmentController) Register(rg *gin.RouterGroup) {
	rg.GET("/service-environments", ctl.list)
	rg.POST("/service-environments", ctl.create)
	rg.DELETE("/service-environments", ctl.delete)
}

// @Summary 列出服务环境绑定
// @Tags service-environments
// @Security BearerAuth
// @Produce json
// @Param service_id query int true "Service ID; invalid -> 400"
// @Param page query int false "Page (>=1); invalid -> 400"
// @Param page_size query int false "Page size (1~100); invalid -> 400"
// @Success 200 {object} dtoresp.ServiceEnvironmentListResponse
// @Failure 400 {object} ErrorResponse
// @Router /service-environments [get]
func (ctl *ServiceEnvironmentController) list(c *gin.Context) {
	serviceIDStr := c.Query("service_id")
	if serviceIDStr == "" {
		failInvalidQuery(c, "service_id", "required")
		return
	}
	serviceID, err := parseUint64(serviceIDStr)
	if err != nil {
		failInvalidQuery(c, "service_id", "invalid")
		return
	}
	page, pageSize, ok := parsePaginationStrict(c)
	if !ok {
		return
	}
	list, total, err := ctl.svc.List(c.Request.Context(), serviceID, page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.ServiceEnvironmentListData{
		List: list,
		Pagination: dtoresp.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	})
}

// @Summary 创建服务环境绑定
// @Tags service-environments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body request.ServiceEnvironmentCreateRequest true "ServiceEnvironmentCreateRequest"
// @Success 200 {object} dtoresp.EmptyResponse
// @Router /service-environments [post]
func (ctl *ServiceEnvironmentController) create(c *gin.Context) {
	var req request.ServiceEnvironmentCreateRequest
	if !bindJSONStrict(c, &req) {
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	if err := ctl.svc.Create(c.Request.Context(), actor, req.ServiceID, req.EnvironmentID); err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.EmptyData{})
}

// @Summary 删除服务环境绑定
// @Tags service-environments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body request.ServiceEnvironmentDeleteRequest true "ServiceEnvironmentDeleteRequest"
// @Success 200 {object} dtoresp.EmptyResponse
// @Router /service-environments [delete]
func (ctl *ServiceEnvironmentController) delete(c *gin.Context) {
	var req request.ServiceEnvironmentDeleteRequest
	if !bindJSONStrict(c, &req) {
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	if err := ctl.svc.Delete(c.Request.Context(), actor, req.ServiceID, req.EnvironmentID); err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.EmptyData{})
}
