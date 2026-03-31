package dtoresp

import "irms/backend/internal/vo"

type EnvironmentListResponse struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Data      EnvironmentListData `json:"data"`
	RequestID string            `json:"request_id"`
}

type EnvironmentListData struct {
	List       []vo.EnvironmentVO `json:"list"`
	Pagination Pagination         `json:"pagination"`
}

type LocationListResponse struct {
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	Data      LocationListData `json:"data"`
	RequestID string         `json:"request_id"`
}

type LocationListData struct {
	List       []vo.LocationVO `json:"list"`
	Pagination Pagination      `json:"pagination"`
}
