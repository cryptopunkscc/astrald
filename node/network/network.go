package network

import (
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/events"
)

type Network interface {
	net.Router
	Events() *events.Queue
	Server() *Server
	AddLink(net.Link) error
	Links() *LinkSet
}
