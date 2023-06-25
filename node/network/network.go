package network

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/events"
)

type Network interface {
	net.Router
	Link(context.Context, id.Identity) (net.Link, error)
	Events() *events.Queue
	Server() *Server
	AddLink(net.Link) error
	Links() *LinkSet
}
