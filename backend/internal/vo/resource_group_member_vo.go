package vo

import "time"

type ResourceGroupMemberVO struct {
	ID         uint64    `json:"id"`
	GroupID    uint64    `json:"group_id"`
	GroupType  string    `json:"group_type"`
	MemberID   uint64    `json:"member_id"`
	MemberType string    `json:"member_type"`
	MemberName string    `json:"member_name"`
	CreatedAt  time.Time `json:"created_at"`
}
