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

var _ dtoresp.HostEnvironmentListResponse

type HostEnvironmentController struct {
	svc *service.HostEnvironmentService
}

func NewHostEnvironmentController(svc *service.HostEnvironmentService) *HostEnvironmentController {
	return &HostEnvironmentController{svc: svc}
}

func (ctl *HostEnvironmentController) Register(rg *gin.RouterGroup) {
	rg.GET("/host-environments", ctl.list)
	rg.POST("/host-environments", ctl.create)
	rg.DELETE("/host-environments", ctl.delete)
}

// @Summary 列出主机环境绑定
// @Tags host-environments
// @Security BearerAuth
// @Produce json
// @Param host_id query int true "Host ID; invalid -> 400"
// @Param page query int false "Page (>=1); invalid -> 400"
// @Param page_size query int false "Page size (1~100); invalid -> 400"
// @Success 200 {object} dtoresp.HostEnvironmentListResponse
// @Failure 400 {object} ErrorResponse
// @Router /host-environments [get]
func (ctl *HostEnvironmentController) list(c *gin.Context) {
	hostIDStr := c.Query("host_id")
	if hostIDStr == "" {
		failInvalidQuery(c, "host_id", "required")
		return
	}
	hostID, err := parseUint64(hostIDStr)
	if err != nil {
		failInvalidQuery(c, "host_id", "invalid")
		return
	}
	page, pageSize, ok := parsePaginationStrict(c)
	if !ok {
		return
	}
	list, total, err := ctl.svc.List(c.Request.Context(), hostID, page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.HostEnvironmentListData{
		List: list,
		Pagination: dtoresp.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	})
}

// @Summary 创建主机环境绑定
// @Tags host-environments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body request.HostEnvironmentCreateRequest true "HostEnvironmentCreateRequest"
// @Success 200 {object} dtoresp.EmptyResponse
// @Router /host-environments [post]
func (ctl *HostEnvironmentController) create(c *gin.Context) {
	var req request.HostEnvironmentCreateRequest
	if !bindJSONStrict(c, &req) {
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	if err := ctl.svc.Create(c.Request.Context(), actor, req.HostID, req.EnvironmentID); err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.EmptyData{})
}

// @Summary 删除主机环境绑定
// @Tags host-environments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body request.HostEnvironmentDeleteRequest true "HostEnvironmentDeleteRequest"
// @Success 200 {object} dtoresp.EmptyResponse
// @Router /host-environments [delete]
func (ctl *HostEnvironmentController) delete(c *gin.Context) {
	var req request.HostEnvironmentDeleteRequest
	if !bindJSONStrict(c, &req) {
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	if err := ctl.svc.Delete(c.Request.Context(), actor, req.HostID, req.EnvironmentID); err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.EmptyData{})
}
