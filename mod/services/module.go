package services

import "github.com/cryptopunkscc/astrald/astral"

const ModuleName = "services"

const (
	MethodServiceDiscovery = "services.discovery"
)

type Module interface {
	AddServiceDiscoverer(ServiceDiscoverer)
	DiscoverRemoteServices(ctx *astral.Context, target *astral.Identity, subscribe bool) error
}

type ServiceDiscoverer interface {
	DiscoverService(ctx *astral.Context, collector *astral.Identity) (snapshot *Service, ch <-chan ServiceChange, err error)
}
