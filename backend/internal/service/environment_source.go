package service

type EnvironmentSource string

const (
	EnvironmentSourceService       EnvironmentSource = "service"
	EnvironmentSourceHost          EnvironmentSource = "host"
	EnvironmentSourceHostInherited EnvironmentSource = "host_inherited"
	EnvironmentSourceNone          EnvironmentSource = "none"
)

