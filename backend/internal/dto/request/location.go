package request

import "strings"

type LocationCreateRequest struct {
	Code         string `json:"code" binding:"required"`
	Name         string `json:"name" binding:"required"`
	LocationType string `json:"location_type" binding:"required"`
	Address      string `json:"address"`
	Status       string `json:"status"`
	Description  string `json:"description"`
}

func (r LocationCreateRequest) StatusOrDefault() string {
	if strings.TrimSpace(r.Status) == "" {
		return "active"
	}
	return r.Status
}

func (r LocationCreateRequest) AddressPtr() *string {
	if strings.TrimSpace(r.Address) == "" {
		return nil
	}
	v := r.Address
	return &v
}

func (r LocationCreateRequest) DescriptionPtr() *string {
	if strings.TrimSpace(r.Description) == "" {
		return nil
	}
	v := r.Description
	return &v
}

type LocationUpdateRequest struct {
	Code         string `json:"code" binding:"required"`
	Name         string `json:"name" binding:"required"`
	LocationType string `json:"location_type" binding:"required"`
	Address      string `json:"address"`
	Status       string `json:"status" binding:"required"`
	Description  string `json:"description"`
}

func (r LocationUpdateRequest) AddressPtr() *string {
	if strings.TrimSpace(r.Address) == "" {
		return nil
	}
	v := r.Address
	return &v
}

func (r LocationUpdateRequest) DescriptionPtr() *string {
	if strings.TrimSpace(r.Description) == "" {
		return nil
	}
	v := r.Description
	return &v
}

