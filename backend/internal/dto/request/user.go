package request

type UserCreateRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role" binding:"omitempty,oneof=super_admin user"`
	Status   string `json:"status" binding:"omitempty,oneof=enabled disabled"`
}

type UserUpdateRequest struct {
	Role   string `json:"role" binding:"required,oneof=super_admin user"`
	Status string `json:"status" binding:"required,oneof=enabled disabled"`
}
