package network

import (
	"github.com/cryptopunkscc/astrald/node/link"
)

type EventPeerLinked struct {
	Peer *Peer
	Link *link.Link
}

type EventPeerUnlinked struct {
	Peer *Peer
}
