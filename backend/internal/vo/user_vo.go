package vo

import "time"

type UserVO struct {
	ID               uint64    `json:"id"`
	Username         string    `json:"username"`
	Role             string    `json:"role"`
	Status           string    `json:"status"`
	MustChangePassword bool    `json:"must_change_password"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

