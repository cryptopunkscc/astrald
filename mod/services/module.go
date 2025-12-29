package services

const ModuleName = "services"

type Module interface {
	AddServiceDiscoverer(ServiceDiscoverer) error
}

type ServiceDiscoverer interface {
}
