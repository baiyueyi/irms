package request

import "strings"

type EnvironmentCreateRequest struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

func (r EnvironmentCreateRequest) StatusOrDefault() string {
	if strings.TrimSpace(r.Status) == "" {
		return "active"
	}
	return r.Status
}

func (r EnvironmentCreateRequest) DescriptionPtr() *string {
	if strings.TrimSpace(r.Description) == "" {
		return nil
	}
	v := r.Description
	return &v
}

type EnvironmentUpdateRequest struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Status      string `json:"status" binding:"required"`
	Description string `json:"description"`
}

func (r EnvironmentUpdateRequest) DescriptionPtr() *string {
	if strings.TrimSpace(r.Description) == "" {
		return nil
	}
	v := r.Description
	return &v
}

