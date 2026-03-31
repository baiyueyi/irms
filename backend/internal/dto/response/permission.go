package dtoresp

import "irms/backend/internal/vo"

type PermissionResourceListResponse struct {
	Code      string                    `json:"code"`
	Message   string                    `json:"message"`
	Data      PermissionResourceListData `json:"data"`
	RequestID string                    `json:"request_id"`
}

type PermissionResourceListData struct {
	List []vo.PermissionResourceVO `json:"list"`
}
