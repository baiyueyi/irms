package vo

import "time"

type HostEnvironmentVO struct {
	ID            uint64    `json:"id"`
	HostID        uint64    `json:"host_id"`
	EnvironmentID uint64    `json:"environment_id"`
	CreatedAt     time.Time `json:"created_at"`
}

