package dtoresp

import (
	"irms/backend/internal/vo"
)

type HostListResponse struct {
	Code      string     `json:"code"`
	Message   string     `json:"message"`
	Data      HostListData `json:"data"`
	RequestID string     `json:"request_id"`
}

type HostListData struct {
	List       []vo.HostVO `json:"list"`
	Pagination Pagination  `json:"pagination"`
}

type HostGetResponse struct {
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Data      vo.HostVO `json:"data"`
	RequestID string    `json:"request_id"`
}
