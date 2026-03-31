package dtoresp

type HealthData struct {
	Status string `json:"status"`
}

type HealthResponse struct {
	Code      string     `json:"code"`
	Message   string     `json:"message"`
	Data      HealthData `json:"data"`
	RequestID string     `json:"request_id"`
}
