package dtoresp

import "irms/backend/internal/vo"

type ResourceListResponse struct {
	Code      string          `json:"code"`
	Message   string          `json:"message"`
	Data      ResourceListData `json:"data"`
	RequestID string          `json:"request_id"`
}

type ResourceListData struct {
	List       []vo.ResourceVO `json:"list"`
	Pagination Pagination      `json:"pagination"`
}
