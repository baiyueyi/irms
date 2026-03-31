package request

type PageCreateRequest struct {
	Name        string `json:"name" binding:"required"`
	RoutePath   string `json:"route_path" binding:"required"`
	Source      string `json:"source"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

type PageUpdateRequest struct {
	Name        string `json:"name" binding:"required"`
	RoutePath   string `json:"route_path" binding:"required"`
	Source      string `json:"source"`
	Status      string `json:"status" binding:"required"`
	Description string `json:"description"`
}

type PageSyncRouteItem struct {
	Name      string `json:"name"`
	RoutePath string `json:"route_path"`
	Source    string `json:"source"`
}

type PageSyncRequest struct {
	DryRun bool              `json:"dry_run"`
	Routes []PageSyncRouteItem `json:"routes"`
}

