package dtoresp

import "irms/backend/internal/vo"

type ResourceGroupListResponse struct {
	Code      string              `json:"code"`
	Message   string              `json:"message"`
	Data      ResourceGroupListData `json:"data"`
	RequestID string              `json:"request_id"`
}

type ResourceGroupListData struct {
	List       []vo.ResourceGroupVO `json:"list"`
	Pagination Pagination           `json:"pagination"`
}

type ResourceGroupMemberListResponse struct {
	Code      string                    `json:"code"`
	Message   string                    `json:"message"`
	Data      ResourceGroupMemberListData `json:"data"`
	RequestID string                    `json:"request_id"`
}

type ResourceGroupMemberListData struct {
	List       []vo.ResourceGroupMemberVO `json:"list"`
	Pagination Pagination                 `json:"pagination"`
}
