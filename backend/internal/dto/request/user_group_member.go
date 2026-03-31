package request

type UserGroupMemberCreateRequest struct {
	UserID      uint64 `json:"user_id" binding:"required"`
	UserGroupID uint64 `json:"user_group_id" binding:"required"`
}

type UserGroupMemberDeleteRequest struct {
	UserID      uint64 `json:"user_id" binding:"required"`
	UserGroupID uint64 `json:"user_group_id" binding:"required"`
}

