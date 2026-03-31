package dtoresp

import "irms/backend/internal/vo"

type UserGroupListResponse struct {
	Code      string           `json:"code"`
	Message   string           `json:"message"`
	Data      UserGroupListData `json:"data"`
	RequestID string           `json:"request_id"`
}

type UserGroupListData struct {
	List       []vo.UserGroupVO `json:"list"`
	Pagination Pagination       `json:"pagination"`
}

type UserGroupMemberListResponse struct {
	Code      string                `json:"code"`
	Message   string                `json:"message"`
	Data      UserGroupMemberListData `json:"data"`
	RequestID string                `json:"request_id"`
}

type UserGroupMemberListData struct {
	List       []vo.UserGroupMemberVO `json:"list"`
	Pagination Pagination             `json:"pagination"`
}

