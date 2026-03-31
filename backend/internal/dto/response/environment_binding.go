package dtoresp

import "irms/backend/internal/vo"

type HostEnvironmentListResponse struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Data      HostEnvironmentListData `json:"data"`
	RequestID string                 `json:"request_id"`
}

type HostEnvironmentListData struct {
	List       []vo.HostEnvironmentVO `json:"list"`
	Pagination Pagination             `json:"pagination"`
}

type ServiceEnvironmentListResponse struct {
	Code      string                    `json:"code"`
	Message   string                    `json:"message"`
	Data      ServiceEnvironmentListData `json:"data"`
	RequestID string                    `json:"request_id"`
}

type ServiceEnvironmentListData struct {
	List       []vo.ServiceEnvironmentVO `json:"list"`
	Pagination Pagination                `json:"pagination"`
}

