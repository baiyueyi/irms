package request

type HostEnvironmentCreateRequest struct {
	HostID        uint64 `json:"host_id" binding:"required"`
	EnvironmentID uint64 `json:"environment_id" binding:"required"`
}

type HostEnvironmentDeleteRequest struct {
	HostID        uint64 `json:"host_id" binding:"required"`
	EnvironmentID uint64 `json:"environment_id" binding:"required"`
}

type ServiceEnvironmentCreateRequest struct {
	ServiceID     uint64 `json:"service_id" binding:"required"`
	EnvironmentID uint64 `json:"environment_id" binding:"required"`
}

type ServiceEnvironmentDeleteRequest struct {
	ServiceID     uint64 `json:"service_id" binding:"required"`
	EnvironmentID uint64 `json:"environment_id" binding:"required"`
}

