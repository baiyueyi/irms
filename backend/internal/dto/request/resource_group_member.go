package request

type ResourceGroupMemberCreateRequest struct {
	GroupID         uint64 `json:"group_id"`
	ResourceGroupID uint64 `json:"resource_group_id"`
	MemberID        uint64 `json:"member_id"`
	ResourceKey     uint64 `json:"resource_key"`
	MemberType      string `json:"member_type"`
	ResourceType    string `json:"resource_type"`
	MemberName      string `json:"member_name"`
	ResourceName    string `json:"resource_name"`
}

type ResourceGroupMemberDeleteRequest struct {
	GroupID         uint64 `json:"group_id"`
	ResourceGroupID uint64 `json:"resource_group_id"`
	MemberID        uint64 `json:"member_id"`
	ResourceKey     uint64 `json:"resource_key"`
	MemberType      string `json:"member_type"`
	ResourceType    string `json:"resource_type"`
	MemberName      string `json:"member_name"`
	ResourceName    string `json:"resource_name"`
}
