package vo

import "time"

type ServiceVO struct {
	ID                   uint64    `json:"id"`
	Name                 string    `json:"name"`
	ServiceKind          string    `json:"service_kind"`
	HostID               *uint64   `json:"host_id"`
	Host                 string    `json:"host"`
	EndpointOrIdentifier string    `json:"endpoint_or_identifier"`
	Port                 *int      `json:"port"`
	Protocol             string    `json:"protocol"`
	CloudVendor          string    `json:"cloud_vendor"`
	CloudProductCode     string    `json:"cloud_product_code"`
	Status               string    `json:"status"`
	Description          string    `json:"description"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`

	OwnEnvironmentIDs      []uint64 `json:"own_environment_ids"`
	InheritedEnvironmentIDs []uint64 `json:"inherited_environment_ids"`
	EnvironmentIDs         []uint64 `json:"environment_ids"`

	Environments          []string `json:"environments"`
	SelfEnvironments      []string `json:"self_environments"`
	EnvironmentSource     string   `json:"environment_source"`
}
