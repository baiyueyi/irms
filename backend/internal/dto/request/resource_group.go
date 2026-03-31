package request

type ResourceGroupCreateRequest struct {
	Name        string `json:"name" binding:"required"`
	GroupType   string `json:"group_type"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type ResourceGroupUpdateRequest struct {
	Name        string `json:"name" binding:"required"`
	GroupType   string `json:"group_type"`
	Type        string `json:"type"`
	Description string `json:"description"`
}
