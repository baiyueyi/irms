package controller

import (
	"strings"

	dtoresp "irms/backend/internal/dto/response"
	apiresp "irms/backend/internal/pkg/response"
	"irms/backend/internal/service"

	"github.com/gin-gonic/gin"
)

var _ dtoresp.AuditLogListResponse

type AuditLogController struct {
	svc *service.AuditService
}

func NewAuditLogController(svc *service.AuditService) *AuditLogController {
	return &AuditLogController{svc: svc}
}

func (ctl *AuditLogController) Register(rg *gin.RouterGroup) {
	rg.GET("/audit-logs", ctl.list)
}

// @Summary 列出审计日志
// @Tags audit-logs
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page (>=1); invalid -> 400"
// @Param page_size query int false "Page size (1~100); invalid -> 400"
// @Param target_type query string false "Target type"
// @Param target_id query string false "Target ID"
// @Param result query string false "Result"
// @Param from query string false "From time"
// @Param to query string false "To time"
// @Param actor_user_id query int false "Actor user ID; invalid -> 400"
// @Success 200 {object} dtoresp.AuditLogListResponse
// @Failure 400 {object} ErrorResponse
// @Router /audit-logs [get]
func (ctl *AuditLogController) list(c *gin.Context) {
	page, pageSize, ok := parsePaginationStrict(c)
	if !ok {
		return
	}
	filter := service.AuditListFilter{
		TargetType: strings.TrimSpace(c.Query("target_type")),
		TargetID:   strings.TrimSpace(c.Query("target_id")),
		Result:     strings.TrimSpace(c.Query("result")),
		From:       strings.TrimSpace(c.Query("from")),
		To:         strings.TrimSpace(c.Query("to")),
	}
	if v := strings.TrimSpace(c.Query("actor_user_id")); v != "" {
		id, err := parseUint64(v)
		if err != nil {
			failInvalidQuery(c, "actor_user_id", "invalid")
			return
		}
		filter.ActorUserID = &id
	}
	list, total, err := ctl.svc.ListPaged(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.AuditLogListData{
		List: list,
		Pagination: dtoresp.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	})
}
