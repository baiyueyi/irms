package request

type ServiceCreateRequest struct {
	Name                 string  `json:"name" binding:"required"`
	ServiceKind          string  `json:"service_kind" binding:"required"`
	HostID               *uint64 `json:"host_id"`
	EndpointOrIdentifier string  `json:"endpoint_or_identifier" binding:"required"`
	Port                 *int    `json:"port"`
	Protocol             string  `json:"protocol"`
	CloudVendor          string  `json:"cloud_vendor"`
	CloudProductCode     string  `json:"cloud_product_code"`
	Status               string  `json:"status"`
	Description          string  `json:"description"`

	EnvironmentIDs *[]uint64 `json:"environment_ids" binding:"required"`
}

type ServiceUpdateRequest struct {
	Name                 string  `json:"name" binding:"required"`
	ServiceKind          string  `json:"service_kind" binding:"required"`
	HostID               *uint64 `json:"host_id"`
	EndpointOrIdentifier string  `json:"endpoint_or_identifier" binding:"required"`
	Port                 *int    `json:"port"`
	Protocol             string  `json:"protocol"`
	CloudVendor          string  `json:"cloud_vendor"`
	CloudProductCode     string  `json:"cloud_product_code"`
	Status               string  `json:"status"`
	Description          string  `json:"description"`

	EnvironmentIDs *[]uint64 `json:"environment_ids" binding:"required"`
}
