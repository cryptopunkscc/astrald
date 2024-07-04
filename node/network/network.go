package network

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/events"
)

type Network interface {
	AddLink(net.Link) error
	Links() *LinkSet
	Events() *events.Queue

	AddLinker(Linker) error
	Linker
}

type Linker interface {
	Link(context.Context, id.Identity) (net.Link, error)
}
