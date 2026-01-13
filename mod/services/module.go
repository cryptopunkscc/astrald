package services

import "github.com/cryptopunkscc/astrald/astral"

const ModuleName = "services"

const (
	MethodServiceDiscovery = "services.discovery"
)

type Module interface {
	AddServiceDiscoverer(ServiceDiscoverer) error
	DiscoverRemoteServices(ctx *astral.Context, target *astral.Identity, subscribe bool) error
}

type DiscoverOptions struct {
	Snapshot bool
	Follow   bool
}

type ServiceDiscoverer interface {
	DiscoverService(
		ctx *astral.Context,
		caller *astral.Identity,
		opts DiscoverOptions,
	) (<-chan ServiceDiscoveryResult, error)
}
