package request

type GrantUpsertRequest struct {
	SubjectType string `json:"subject_type" binding:"required"`
	SubjectID   uint64 `json:"subject_id" binding:"required"`
	ObjectType  string `json:"object_type" binding:"required"`
	ObjectID    uint64 `json:"object_id" binding:"required"`
	Permission  string `json:"permission" binding:"required"`
}

type GrantUpdateRequest struct {
	Permission string `json:"permission" binding:"required"`
}
