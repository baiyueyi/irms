package dtoresp

type IDResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Data      IDData `json:"data"`
	RequestID string `json:"request_id"`
}

type IDData struct {
	ID uint64 `json:"id"`
}

type KeyResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Data      KeyData `json:"data"`
	RequestID string `json:"request_id"`
}

type KeyData struct {
	Key uint64 `json:"key"`
}

type EmptyResponse struct {
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Data      EmptyData `json:"data"`
	RequestID string    `json:"request_id"`
}

type EmptyData struct{}
