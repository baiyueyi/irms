package service

// Deprecated: ResourceListFilter 仅用于 resources compatibility storage。
// 新业务能力禁止围绕 resources/resource_groups/resource_group_members 新增查询语义。
type ResourceListFilter struct {
	Keyword string
	Type    string
	Status  string
}
