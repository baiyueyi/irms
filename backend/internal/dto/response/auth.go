package dtoresp

import "irms/backend/internal/vo"

type LoginData struct {
	Token string   `json:"token"`
	User  vo.UserVO `json:"user"`
}

type LoginResponse struct {
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Data      LoginData `json:"data"`
	RequestID string    `json:"request_id"`
}

type MeResponse struct {
	Code      string   `json:"code"`
	Message   string   `json:"message"`
	Data      vo.UserVO `json:"data"`
	RequestID string   `json:"request_id"`
}

