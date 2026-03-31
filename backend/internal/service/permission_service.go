package service

import (
	"context"

	"irms/backend/internal/repository"
	"irms/backend/internal/store"
	"irms/backend/internal/vo"

)

type PermissionService struct {
	repo *repository.PermissionRepository
}

func NewPermissionService(st *store.Store) *PermissionService {
	return &PermissionService{repo: repository.NewPermissionRepository(st.Gorm)}
}

func (s *PermissionService) ListMyPageResources(ctx context.Context, userID uint64) ([]vo.PermissionResourceVO, error) {
	rows, err := s.repo.ListMyPageResources(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]vo.PermissionResourceVO, 0, len(rows))
	for _, r := range rows {
		out = append(out, vo.PermissionResourceVO{
			ResourceKey:  r.ResourceKey,
			ResourceName: r.ResourceName,
			ResourceType: r.ResourceType,
			RoutePath:    r.RoutePath,
			Permission:   r.Permission,
		})
	}
	return out, nil
}
