package dtoresp

import "irms/backend/internal/vo"

type AuditLogListResponse struct {
	Code      string          `json:"code"`
	Message   string          `json:"message"`
	Data      AuditLogListData `json:"data"`
	RequestID string          `json:"request_id"`
}

type AuditLogListData struct {
	List       []vo.AuditLogVO `json:"list"`
	Pagination Pagination      `json:"pagination"`
}
