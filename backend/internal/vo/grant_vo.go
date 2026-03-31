package vo

import "time"

type GrantVO struct {
	ID                uint64    `json:"id"`
	SubjectType       string    `json:"subject_type"`
	SubjectTypeDisplay string   `json:"subject_type_display"`
	SubjectID         uint64    `json:"subject_id"`
	SubjectName       string    `json:"subject_name"`
	ObjectType        string    `json:"object_type"`
	ObjectTypeDisplay string    `json:"object_type_display"`
	ObjectID          uint64    `json:"object_id"`
	ObjectName        string    `json:"object_name"`
	Permission        string    `json:"permission"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
