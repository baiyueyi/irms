package dtoresp

import (
	"irms/backend/internal/vo"
)

type ServiceListResponse struct {
	Code      string       `json:"code"`
	Message   string       `json:"message"`
	Data      ServiceListData `json:"data"`
	RequestID string       `json:"request_id"`
}

type ServiceListData struct {
	List       []vo.ServiceVO `json:"list"`
	Pagination Pagination     `json:"pagination"`
}

type ServiceGetResponse struct {
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	Data      vo.ServiceVO `json:"data"`
	RequestID string      `json:"request_id"`
}
