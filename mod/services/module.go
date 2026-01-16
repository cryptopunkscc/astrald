package services

import "github.com/cryptopunkscc/astrald/astral"

const ModuleName = "services"
const DBPrefix = "services__"

type Module interface {
	AddDiscoverer(Discoverer) error
	Discoverer
}

type Discoverer interface {
	DiscoverServices(ctx *astral.Context, caller *astral.Identity, follow bool) (<-chan *Update, error)
}
