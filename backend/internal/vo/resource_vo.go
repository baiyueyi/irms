package vo

import "time"

type ResourceVO struct {
	Key               uint64    `json:"key"`
	Name              string    `json:"name"`
	Type              string    `json:"type"`
	Address           string    `json:"address"`
	ServiceIdentifier string    `json:"service_identifier"`
	RoutePath         string    `json:"route_path"`
	Status            string    `json:"status"`
	Description       string    `json:"description"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
