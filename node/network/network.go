package network

import (
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/events"
)

type Network interface {
	Events() *events.Queue
	Server() *Server
	AddLink(net.Link) error
	Links() *LinkSet
}
