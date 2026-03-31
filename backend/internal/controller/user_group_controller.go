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

var _ dtoresp.UserGroupListResponse
var _ dtoresp.UserGroupMemberListResponse

type UserGroupController struct {
	svc *service.UserGroupService
}

func NewUserGroupController(svc *service.UserGroupService) *UserGroupController {
	return &UserGroupController{svc: svc}
}

func (ctl *UserGroupController) Register(rg *gin.RouterGroup) {
	rg.GET("/user-groups", ctl.listGroups)
	rg.POST("/user-groups", ctl.createGroup)
	rg.PUT("/user-groups/:id", ctl.updateGroup)
	rg.DELETE("/user-groups/:id", ctl.deleteGroup)

	rg.GET("/user-group-members", ctl.listMembers)
	rg.POST("/user-group-members", ctl.addMember)
	rg.DELETE("/user-group-members", ctl.removeMember)
}

// @Summary 列出用户组
// @Tags user-groups
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page (>=1); invalid -> 400"
// @Param page_size query int false "Page size (1~100); invalid -> 400"
// @Param keyword query string false "Search keyword (name like)"
// @Success 200 {object} dtoresp.UserGroupListResponse
// @Failure 400 {object} ErrorResponse
// @Router /user-groups [get]
func (ctl *UserGroupController) listGroups(c *gin.Context) {
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
	apiresp.OK(c, dtoresp.UserGroupListData{
		List: list,
		Pagination: dtoresp.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	})
}

// @Summary 创建用户组
// @Tags user-groups
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body request.UserGroupCreateRequest true "UserGroupCreateRequest"
// @Success 200 {object} dtoresp.IDResponse
// @Router /user-groups [post]
func (ctl *UserGroupController) createGroup(c *gin.Context) {
	var req request.UserGroupCreateRequest
	if !bindJSONStrict(c, &req) {
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	id, err := ctl.svc.Create(c.Request.Context(), actor, req.Name, req.Description)
	if err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.IDData{ID: id})
}

// @Summary 更新用户组
// @Tags user-groups
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "UserGroup ID"
// @Param body body request.UserGroupUpdateRequest true "UserGroupUpdateRequest"
// @Success 200 {object} dtoresp.EmptyResponse
// @Router /user-groups/{id} [put]
func (ctl *UserGroupController) updateGroup(c *gin.Context) {
	id, ok := parseUint64Param(c, "id")
	if !ok {
		return
	}
	var req request.UserGroupUpdateRequest
	if !bindJSONStrict(c, &req) {
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	if err := ctl.svc.Update(c.Request.Context(), actor, id, req.Name, req.Description); err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.EmptyData{})
}

// @Summary 删除用户组
// @Tags user-groups
// @Security BearerAuth
// @Produce json
// @Param id path int true "UserGroup ID"
// @Success 200 {object} dtoresp.EmptyResponse
// @Router /user-groups/{id} [delete]
func (ctl *UserGroupController) deleteGroup(c *gin.Context) {
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

// @Summary 列出用户组成员
// @Tags user-group-members
// @Security BearerAuth
// @Produce json
// @Param user_group_id query int true "UserGroup ID"
// @Param page query int false "Page (>=1); invalid -> 400"
// @Param page_size query int false "Page size (1~100); invalid -> 400"
// @Param keyword query string false "Search keyword (username like)"
// @Success 200 {object} dtoresp.UserGroupMemberListResponse
// @Failure 400 {object} ErrorResponse
// @Router /user-group-members [get]
func (ctl *UserGroupController) listMembers(c *gin.Context) {
	groupIDStr := strings.TrimSpace(c.Query("user_group_id"))
	if groupIDStr == "" {
		failInvalidQuery(c, "user_group_id", "required")
		return
	}
	groupID, err := parseUint64(groupIDStr)
	if err != nil {
		failInvalidQuery(c, "user_group_id", "invalid")
		return
	}
	page, pageSize, ok := parsePaginationStrict(c)
	if !ok {
		return
	}
	keyword := strings.TrimSpace(c.Query("keyword"))
	list, total, err := ctl.svc.ListMembers(c.Request.Context(), groupID, keyword, page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.UserGroupMemberListData{
		List: list,
		Pagination: dtoresp.Pagination{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	})
}

// @Summary 添加用户组成员
// @Tags user-group-members
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body request.UserGroupMemberCreateRequest true "UserGroupMemberCreateRequest"
// @Success 200 {object} dtoresp.EmptyResponse
// @Router /user-group-members [post]
func (ctl *UserGroupController) addMember(c *gin.Context) {
	var req request.UserGroupMemberCreateRequest
	if !bindJSONStrict(c, &req) {
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	if err := ctl.svc.AddMember(c.Request.Context(), actor, req.UserID, req.UserGroupID); err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.EmptyData{})
}

// @Summary 移除用户组成员
// @Tags user-group-members
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body request.UserGroupMemberDeleteRequest true "UserGroupMemberDeleteRequest"
// @Success 200 {object} dtoresp.EmptyResponse
// @Router /user-group-members [delete]
func (ctl *UserGroupController) removeMember(c *gin.Context) {
	var req request.UserGroupMemberDeleteRequest
	if !bindJSONStrict(c, &req) {
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		apiresp.Fail(c, http.StatusUnauthorized, ecode.CodeUnauthorized, "unauthorized", nil)
		return
	}
	if err := ctl.svc.RemoveMember(c.Request.Context(), actor, req.UserID, req.UserGroupID); err != nil {
		handleError(c, err)
		return
	}
	apiresp.OK(c, dtoresp.EmptyData{})
}
