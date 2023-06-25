package network

import (
	"github.com/cryptopunkscc/astrald/net"
)

type EventLinkAdded struct {
	Link net.Link
}

type EventLinkRemoved struct {
	Link net.Link
}
