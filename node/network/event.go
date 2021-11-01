package network

import "github.com/cryptopunkscc/astrald/astral/link"

const (
	EventLinkUp       = "network.link_up"
	EventLinkDown     = "network.link_down"
	EventPeerLinked   = "network.peer_linked"
	EventPeerUnlinked = "network.peer_unlinked"
)

type Event struct {
	Type string
	Link *link.Link
	Peer *Peer
}

func (e Event) Event() string {
	return e.Type
}
