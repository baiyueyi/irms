package vo

import "time"

type ServiceEnvironmentVO struct {
	ID            uint64    `json:"id"`
	ServiceID     uint64    `json:"service_id"`
	EnvironmentID uint64    `json:"environment_id"`
	CreatedAt     time.Time `json:"created_at"`
}

