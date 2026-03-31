package vo

type PermissionResourceVO struct {
	ResourceKey  uint64 `json:"resource_key"`
	ResourceName string `json:"resource_name"`
	ResourceType string `json:"resource_type"`
	RoutePath    string `json:"route_path"`
	Permission   string `json:"permission"`
}
