package vo

import "time"

type UserGroupMemberVO struct {
	ID          uint64    `json:"id"`
	UserID      uint64    `json:"user_id"`
	UserGroupID uint64    `json:"user_group_id"`
	UserName    string    `json:"user_name"`
	CreatedAt   time.Time `json:"created_at"`
}

