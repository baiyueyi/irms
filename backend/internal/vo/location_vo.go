package vo

import "time"

type LocationVO struct {
	ID           uint64    `json:"id"`
	Code         string    `json:"code"`
	Name         string    `json:"name"`
	LocationType string    `json:"location_type"`
	Address      string    `json:"address"`
	Status       string    `json:"status"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

