package controller

import (
	"net/http"

	dtoresp "irms/backend/internal/dto/response"
	ecode "irms/backend/internal/pkg/errors"
	apiresp "irms/backend/internal/pkg/response"
	"irms/backend/internal/service"

	"github.com/gin-gonic/gin"
)

var _ dtoresp.PermissionResourceListResponse

type PermissionController struct {
	svc *service.PermissionService
}

func NewPermissionController(svc *service.PermissionService) *PermissionController {
	return &PermissionController{svc: svc}
}

func (ctl *PermissionController) Register(rg *gin.RouterGroup) {
	rg.GET("/permissions/resources", ctl.listMyPageResources)
}

// @Summary 列出我的可访问页面（兼容路径：/permissions/resources）
// @Tags permissions
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dtoresp.PermissionResourceListResponse
// @Router /permissions/resources [get]
func (ctl *PermissionController) listMyPageResources(c *gin.Context) {
	actor, ok := currentActor(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	list, err := ctl.svc.ListMyPageResources(c.Request.Context(), actor.UserID)
	if err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.PermissionResourceListData{List: list})
}
