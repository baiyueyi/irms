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

var _ dtoresp.ResourceGroupListResponse
var _ dtoresp.ResourceGroupMemberListResponse

type ResourceGroupController struct {
	svc *service.ResourceGroupService
}

func NewResourceGroupController(svc *service.ResourceGroupService) *ResourceGroupController {
	return &ResourceGroupController{svc: svc}
}

func (ctl *ResourceGroupController) Register(rg *gin.RouterGroup) {
	rg.GET("/resource-groups", ctl.list)
	rg.POST("/resource-groups", ctl.create)
	rg.PUT("/resource-groups/:id", ctl.update)
	rg.DELETE("/resource-groups/:id", ctl.delete)

	rg.GET("/resource-group-members", ctl.listMembers)
	rg.POST("/resource-group-members", ctl.addMember)
	rg.DELETE("/resource-group-members", ctl.removeMember)
}

// @Summary 列出主机组/服务组
// @Tags resource-groups
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page (>=1); invalid -> 400"
// @Param page_size query int false "Page size (1~100); invalid -> 400"
// @Param keyword query string false "Search keyword (name like)"
// @Param group_type query string false "Group type; invalid -> 400" Enums(host_group,service_group)
// @Param type query string false "Deprecated: use group_type" Enums(host,service)
// @Success 200 {object} dtoresp.ResourceGroupListResponse
// @Failure 400 {object} ErrorResponse
// @Router /resource-groups [get]
func (ctl *ResourceGroupController) list(c *gin.Context) {
	page, pageSize, ok := parsePaginationStrict(c)
	if !ok {
		return
	}
	groupType, ok := normalizeGroupTypeInput(strings.TrimSpace(c.Query("group_type")), strings.TrimSpace(c.Query("type")))
	if !ok {
		failInvalidQuery(c, "group_type", "must be one of: host_group, service_group")
		return
	}
	filter := service.ResourceGroupListFilter{
		Keyword: strings.TrimSpace(c.Query("keyword")),
		Type:    groupType,
	}
	list, total, err := ctl.svc.ListPaged(c.Request.Context(), filter, page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.ResourceGroupListData{
		List: list,
		Pagination: dtoresp.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	})
}

// @Summary 创建资源组
// @Tags resource-groups
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body request.ResourceGroupCreateRequest true "group_type 为主，type 为兼容字段"
// @Success 200 {object} dtoresp.IDResponse
// @Router /resource-groups [post]
func (ctl *ResourceGroupController) create(c *gin.Context) {
	var req request.ResourceGroupCreateRequest
	if !bindJSONStrict(c, &req) {
		return
	}
	groupType, ok := normalizeGroupTypeInput(req.GroupType, req.Type)
	if !ok {
		apiresp.Fail(c, http.StatusBadRequest, ecode.CodeInvalidArgument, "invalid body", map[string]any{"field": "group_type", "message": "must be one of: host_group, service_group"})
		return
	}
	if groupType == "" {
		apiresp.Fail(c, http.StatusBadRequest, ecode.CodeInvalidArgument, "invalid body", map[string]any{"field": "group_type", "tag": "required"})
		return
	}
	req.Type = groupType
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

// @Summary 更新资源组
// @Tags resource-groups
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Group ID"
// @Param body body request.ResourceGroupUpdateRequest true "group_type 为主，type 为兼容字段"
// @Success 200 {object} dtoresp.EmptyResponse
// @Router /resource-groups/{id} [put]
func (ctl *ResourceGroupController) update(c *gin.Context) {
	id, ok := parseUint64Param(c, "id")
	if !ok {
		return
	}
	var req request.ResourceGroupUpdateRequest
	if !bindJSONStrict(c, &req) {
		return
	}
	groupType, ok := normalizeGroupTypeInput(req.GroupType, req.Type)
	if !ok {
		apiresp.Fail(c, http.StatusBadRequest, ecode.CodeInvalidArgument, "invalid body", map[string]any{"field": "group_type", "message": "must be one of: host_group, service_group"})
		return
	}
	if groupType == "" {
		apiresp.Fail(c, http.StatusBadRequest, ecode.CodeInvalidArgument, "invalid body", map[string]any{"field": "group_type", "tag": "required"})
		return
	}
	req.Type = groupType
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

// @Summary 删除资源组
// @Tags resource-groups
// @Security BearerAuth
// @Produce json
// @Param id path int true "Group ID"
// @Success 200 {object} dtoresp.EmptyResponse
// @Router /resource-groups/{id} [delete]
func (ctl *ResourceGroupController) delete(c *gin.Context) {
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

// @Summary 列出主机组/服务组成员
// @Tags resource-groups
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page (>=1); invalid -> 400"
// @Param page_size query int false "Page size (1~100); invalid -> 400"
// @Param group_id query int false "Group ID; invalid -> 400"
// @Param group_type query string false "Group type; invalid -> 400" Enums(host_group,service_group)
// @Param resource_group_id query int false "Deprecated: use group_id"
// @Param keyword query string false "Search keyword"
// @Success 200 {object} dtoresp.ResourceGroupMemberListResponse
// @Failure 400 {object} ErrorResponse
// @Router /resource-group-members [get]
func (ctl *ResourceGroupController) listMembers(c *gin.Context) {
	page, pageSize, ok := parsePaginationStrict(c)
	if !ok {
		return
	}
	groupIDStr := strings.TrimSpace(c.Query("group_id"))
	if groupIDStr == "" {
		groupIDStr = strings.TrimSpace(c.Query("resource_group_id"))
	}
	if groupIDStr == "" {
		failInvalidQuery(c, "group_id", "required")
		return
	}
	groupID, err := parseUint64(groupIDStr)
	if err != nil || groupID == 0 {
		failInvalidQuery(c, "group_id", "invalid")
		return
	}
	groupType, ok := normalizeGroupTypeInput(strings.TrimSpace(c.Query("group_type")), "")
	if !ok {
		failInvalidQuery(c, "group_type", "must be one of: host_group, service_group")
		return
	}
	keyword := strings.TrimSpace(c.Query("keyword"))
	list, total, err := ctl.svc.ListMembersPaged(c.Request.Context(), groupID, groupType, keyword, page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.ResourceGroupMemberListData{
		List: list,
		Pagination: dtoresp.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	})
}

// @Summary 添加资源组成员
// @Tags resource-groups
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body request.ResourceGroupMemberCreateRequest true "group_id/member_id/member_type 为主，resource_* 为兼容字段"
// @Success 200 {object} dtoresp.EmptyResponse
// @Router /resource-group-members [post]
func (ctl *ResourceGroupController) addMember(c *gin.Context) {
	var req request.ResourceGroupMemberCreateRequest
	if !bindJSONStrict(c, &req) {
		return
	}
	groupID := firstUint64(req.GroupID, req.ResourceGroupID)
	if groupID == 0 {
		apiresp.Fail(c, http.StatusBadRequest, ecode.CodeInvalidArgument, "invalid body", map[string]any{"field": "group_id", "tag": "required"})
		return
	}
	memberID := firstUint64(req.MemberID, req.ResourceKey)
	if memberID == 0 {
		apiresp.Fail(c, http.StatusBadRequest, ecode.CodeInvalidArgument, "invalid body", map[string]any{"field": "member_id", "tag": "required"})
		return
	}
	memberType := strings.TrimSpace(req.MemberType)
	if memberType == "" {
		memberType = strings.TrimSpace(req.ResourceType)
	}
	actor, ok := currentActor(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	if err := ctl.svc.AddMember(c.Request.Context(), actor, memberID, groupID, memberType); err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.EmptyData{})
}

// @Summary 移除资源组成员
// @Tags resource-groups
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body request.ResourceGroupMemberDeleteRequest true "group_id/member_id/member_type 为主，resource_* 为兼容字段"
// @Success 200 {object} dtoresp.EmptyResponse
// @Router /resource-group-members [delete]
func (ctl *ResourceGroupController) removeMember(c *gin.Context) {
	var req request.ResourceGroupMemberDeleteRequest
	if !bindJSONStrict(c, &req) {
		return
	}
	groupID := firstUint64(req.GroupID, req.ResourceGroupID)
	if groupID == 0 {
		apiresp.Fail(c, http.StatusBadRequest, ecode.CodeInvalidArgument, "invalid body", map[string]any{"field": "group_id", "tag": "required"})
		return
	}
	memberID := firstUint64(req.MemberID, req.ResourceKey)
	if memberID == 0 {
		apiresp.Fail(c, http.StatusBadRequest, ecode.CodeInvalidArgument, "invalid body", map[string]any{"field": "member_id", "tag": "required"})
		return
	}
	memberType := strings.TrimSpace(req.MemberType)
	if memberType == "" {
		memberType = strings.TrimSpace(req.ResourceType)
	}
	actor, ok := currentActor(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	if err := ctl.svc.RemoveMember(c.Request.Context(), actor, memberID, groupID, memberType); err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.EmptyData{})
}

func normalizeGroupTypeInput(groupType string, legacyType string) (string, bool) {
	v := strings.TrimSpace(groupType)
	if v == "" {
		v = strings.TrimSpace(legacyType)
	}
	switch v {
	case "":
		return "", true
	case "host_group", "host":
		return "host", true
	case "service_group", "service":
		return "service", true
	default:
		return "", false
	}
}

func firstUint64(values ...uint64) uint64 {
	for _, v := range values {
		if v > 0 {
			return v
		}
	}
	return 0
}
