package request

type HostCreateRequest struct {
	Name            string  `json:"name" binding:"required"`
	Hostname        string  `json:"hostname" binding:"required"`
	PrimaryAddress  string  `json:"primary_address" binding:"required"`
	ProviderKind    string  `json:"provider_kind" binding:"required"`
	CloudVendor     string  `json:"cloud_vendor"`
	CloudInstanceID string  `json:"cloud_instance_id"`
	OSType          string  `json:"os_type"`
	Status          string  `json:"status"`
	LocationID      *uint64 `json:"location_id"`
	Description     string  `json:"description"`

	EnvironmentIDs *[]uint64 `json:"environment_ids" binding:"required"`
}

type HostUpdateRequest struct {
	Name            string   `json:"name" binding:"required"`
	Hostname        string   `json:"hostname" binding:"required"`
	PrimaryAddress  string   `json:"primary_address" binding:"required"`
	ProviderKind    string   `json:"provider_kind" binding:"required"`
	CloudVendor     string   `json:"cloud_vendor"`
	CloudInstanceID string   `json:"cloud_instance_id"`
	OSType          string   `json:"os_type"`
	Status          string   `json:"status"`
	LocationID      *uint64  `json:"location_id"`
	Description     string   `json:"description"`

	EnvironmentIDs *[]uint64 `json:"environment_ids" binding:"required"`
}
