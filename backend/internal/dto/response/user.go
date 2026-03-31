package dtoresp

import "irms/backend/internal/vo"

type UserListResponse struct {
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	Data      UserListData `json:"data"`
	RequestID string      `json:"request_id"`
}

type UserListData struct {
	List       []vo.UserVO `json:"list"`
	Pagination Pagination  `json:"pagination"`
}

