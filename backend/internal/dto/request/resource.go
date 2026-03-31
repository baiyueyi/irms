package request

type ResourceCreateRequest struct {
	Name              string `json:"name" binding:"required"`
	Type              string `json:"type" binding:"required,oneof=host service"`
	Address           string `json:"address"`
	ServiceIdentifier string `json:"service_identifier"`
	RoutePath         string `json:"route_path"`
	Status            string `json:"status" binding:"omitempty,oneof=active inactive"`
	Description       string `json:"description"`
}

type ResourceUpdateRequest struct {
	Name              string `json:"name" binding:"required"`
	Type              string `json:"type" binding:"required,oneof=host service"`
	Address           string `json:"address"`
	ServiceIdentifier string `json:"service_identifier"`
	RoutePath         string `json:"route_path"`
	Status            string `json:"status" binding:"omitempty,oneof=active inactive"`
	Description       string `json:"description"`
}
