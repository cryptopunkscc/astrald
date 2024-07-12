package node

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/events"
	"github.com/cryptopunkscc/astrald/net"
)

type NetworkEngine interface {
	AddLink(net.Link) error
	Links() LinkSet
	Events() *events.Queue

	AddLinker(Linker) error
	Link(context.Context, id.Identity) (net.Link, error)
}

type Linker interface {
	Link(context.Context, id.Identity) (net.Link, error)
}
