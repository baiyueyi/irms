package vo

import "time"

type HostVO struct {
	ID              uint64    `json:"id"`
	Name            string    `json:"name"`
	Hostname        string    `json:"hostname"`
	PrimaryAddress  string    `json:"primary_address"`
	ProviderKind    string    `json:"provider_kind"`
	CloudVendor     string    `json:"cloud_vendor"`
	CloudInstanceID string    `json:"cloud_instance_id"`
	OSType          string    `json:"os_type"`
	Status          string    `json:"status"`
	LocationID      *uint64   `json:"location_id"`
	Location        string    `json:"location"`

	OwnEnvironmentIDs       []uint64 `json:"own_environment_ids"`
	InheritedEnvironmentIDs []uint64 `json:"inherited_environment_ids"`
	EnvironmentIDs  []uint64  `json:"environment_ids"`
	Environments    []string  `json:"environments"`
	Description     string    `json:"description"`
	SelfEnvironments []string `json:"self_environments"`
	EnvironmentSource string  `json:"environment_source"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
