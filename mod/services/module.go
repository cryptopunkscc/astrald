package services

import "github.com/cryptopunkscc/astrald/astral"

const ModuleName = "services"

type Module interface {
	AddServiceDiscoverer(ServiceDiscoverer)
}

type ServiceDiscoverer interface {
	DiscoverService(ctx *astral.Context, collector *astral.Identity) (snapshot *Service, ch <-chan ServiceChange, err error)
}
