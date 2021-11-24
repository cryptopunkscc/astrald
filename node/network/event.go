package network

import (
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/node/network/peer"
)

const (
	EventLinkUp       = "network.link_up"
	EventLinkDown     = "network.link_down"
	EventPeerLinked   = "network.peer_linked"
	EventPeerUnlinked = "network.peer_unlinked"
)

type Event struct {
	Type string
	Link *link.Link
	Peer *peer.Peer
}

func (e Event) Event() string {
	return e.Type
}
