package vo

type PageSyncSummaryVO struct {
	DryRun     bool `json:"dry_run"`
	InputTotal int  `json:"input_total"`
	Created    int  `json:"created"`
	Updated    int  `json:"updated"`
	Unchanged  int  `json:"unchanged"`
	Retired    int  `json:"retired"`
}

type PageSyncRouteVO struct {
	RoutePath string `json:"route_path"`
	Name      string `json:"name"`
	Source    string `json:"source"`
	Status    string `json:"status"`
}

type PageSyncVO struct {
	Summary        PageSyncSummaryVO `json:"summary"`
	NewRoutes      []PageSyncRouteVO `json:"new_routes"`
	ExistingRoutes []PageSyncRouteVO `json:"existing_routes"`
	ChangedRoutes  []PageSyncRouteVO `json:"changed_routes"`
	RetiredRoutes  []PageSyncRouteVO `json:"retired_routes"`
}
