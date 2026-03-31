package service

type HostListFilter struct {
	Keyword       string
	Status        string
	ProviderKind  string
	LocationID    *uint64
	EnvironmentID *uint64
}

