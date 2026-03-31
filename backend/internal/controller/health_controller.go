package controller

import (
	dtoresp "irms/backend/internal/dto/response"
	apiresp "irms/backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

var _ dtoresp.HealthResponse

type HealthController struct{}

func NewHealthController() *HealthController {
	return &HealthController{}
}

// @Summary 健康检查
// @Tags health
// @Produce json
// @Success 200 {object} dtoresp.HealthResponse
// @Router /health [get]
func (ctl *HealthController) Health(c *gin.Context) {
	apiresp.OK(c, dtoresp.HealthData{Status: "ok"})
}
