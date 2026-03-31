package dtoresp

import "irms/backend/internal/vo"

type PageListResponse struct {
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	Data      PageListData `json:"data"`
	RequestID string      `json:"request_id"`
}

type PageListData struct {
	List       []vo.PageVO `json:"list"`
	Pagination Pagination  `json:"pagination"`
}

type PageSyncResponse struct {
	Code      string       `json:"code"`
	Message   string       `json:"message"`
	Data      vo.PageSyncVO `json:"data"`
	RequestID string       `json:"request_id"`
}
