package request

type UserGroupCreateRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UserGroupUpdateRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

