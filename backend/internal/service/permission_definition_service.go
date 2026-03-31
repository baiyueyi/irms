package service

import (
	"context"
	"sort"
	"strings"

	"irms/backend/internal/query"
)

type PermissionDefinitionService struct {
	q *query.Query
}

func NewPermissionDefinitionService(q *query.Query) *PermissionDefinitionService {
	return &PermissionDefinitionService{q: q}
}

func normalizeObjectFamily(objectType string) string {
	switch objectType {
	case "host_group":
		return "host"
	case "service_group":
		return "service"
	default:
		return objectType
	}
}

func (s *PermissionDefinitionService) IsValidPermission(ctx context.Context, objectType string, permissionCode string) (bool, error) {
	code := strings.TrimSpace(strings.ToLower(permissionCode))
	if code == "" {
		return false, nil
	}
	family := normalizeObjectFamily(strings.TrimSpace(strings.ToLower(objectType)))
	qpd := s.q.PermissionDefinition
	cnt, err := qpd.WithContext(ctx).
		Where(qpd.Code.Eq(code), qpd.ObjectFamily.Eq(family), qpd.Status.Eq("active")).
		Count()
	if err != nil {
		return false, err
	}
	return cnt > 0, nil
}

func (s *PermissionDefinitionService) ExistsPermissionCode(ctx context.Context, permissionCode string) (bool, error) {
	code := strings.TrimSpace(strings.ToLower(permissionCode))
	if code == "" {
		return false, nil
	}
	qpd := s.q.PermissionDefinition
	cnt, err := qpd.WithContext(ctx).
		Where(qpd.Code.Eq(code), qpd.Status.Eq("active")).
		Count()
	if err != nil {
		return false, err
	}
	return cnt > 0, nil
}

func (s *PermissionDefinitionService) ExpandPermissionCodes(ctx context.Context, codes []string) ([]string, error) {
	inputSet := map[string]struct{}{}
	for _, code := range codes {
		normalized := strings.TrimSpace(strings.ToLower(code))
		if normalized == "" {
			continue
		}
		inputSet[normalized] = struct{}{}
	}
	if len(inputSet) == 0 {
		return nil, nil
	}
	inputCodes := make([]string, 0, len(inputSet))
	for code := range inputSet {
		inputCodes = append(inputCodes, code)
	}
	qpd := s.q.PermissionDefinition
	inputDefs, err := qpd.WithContext(ctx).
		Where(qpd.Code.In(inputCodes...), qpd.Status.Eq("active")).
		Find()
	if err != nil {
		return nil, err
	}
	codeSet := map[string]struct{}{}
	readFamilies := map[string]struct{}{}
	for _, item := range inputDefs {
		codeSet[item.Code] = struct{}{}
		if item.Action == "read" {
			readFamilies[item.ObjectFamily] = struct{}{}
		}
	}
	if len(readFamilies) > 0 {
		families := make([]string, 0, len(readFamilies))
		for family := range readFamilies {
			families = append(families, family)
		}
		writeDefs, err := qpd.WithContext(ctx).
			Where(qpd.ObjectFamily.In(families...), qpd.Action.Eq("write"), qpd.Status.Eq("active")).
			Find()
		if err != nil {
			return nil, err
		}
		for _, item := range writeDefs {
			codeSet[item.Code] = struct{}{}
		}
	}
	out := make([]string, 0, len(codeSet))
	for code := range codeSet {
		out = append(out, code)
	}
	sort.Strings(out)
	return out, nil
}
