package vo

import "time"

type AuditLogVO struct {
	ID                    uint64    `json:"id"`
	ActorUserID           uint64    `json:"actor_user_id"`
	ActorUsernameSnapshot string    `json:"actor_username_snapshot"`
	Action                string    `json:"action"`
	TargetType            string    `json:"target_type"`
	TargetID              string    `json:"target_id"`
	TargetNameSnapshot    string    `json:"target_name_snapshot"`
	OccurredAt            time.Time `json:"occurred_at"`
	BeforeJSON            string    `json:"before_json"`
	AfterJSON             string    `json:"after_json"`
	Result                string    `json:"result"`
	IP                    string    `json:"ip"`
}
