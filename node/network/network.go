package network

import (
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/events"
)

type Network interface {
	AddLink(net.Link) error
	Links() *LinkSet
	Events() *events.Queue
}
