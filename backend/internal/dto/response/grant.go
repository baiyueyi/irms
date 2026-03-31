package dtoresp

import "irms/backend/internal/vo"

type GrantListResponse struct {
	Code      string       `json:"code"`
	Message   string       `json:"message"`
	Data      GrantListData `json:"data"`
	RequestID string       `json:"request_id"`
}

type GrantListData struct {
	List       []vo.GrantVO `json:"list"`
	Pagination Pagination   `json:"pagination"`
}
