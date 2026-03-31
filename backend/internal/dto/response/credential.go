package dtoresp

import "irms/backend/internal/vo"

type CredentialListResponse struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Data      CredentialListData `json:"data"`
	RequestID string            `json:"request_id"`
}

type CredentialListData struct {
	List       []vo.CredentialVO `json:"list"`
	Pagination Pagination        `json:"pagination"`
}

type CredentialRevealResponse struct {
	Code      string              `json:"code"`
	Message   string              `json:"message"`
	Data      vo.CredentialRevealVO `json:"data"`
	RequestID string              `json:"request_id"`
}
