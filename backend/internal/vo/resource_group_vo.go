package vo

import "time"

type ResourceGroupVO struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	GroupType   string    `json:"group_type"`
	Description string    `json:"description"`
	MemberCount int       `json:"member_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
