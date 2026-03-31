package vo

import "time"

type PageVO struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	RoutePath   string    `json:"route_path"`
	Source      string    `json:"source"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	GrantCount  int       `json:"grant_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

