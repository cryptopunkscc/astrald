package peers

import "github.com/cryptopunkscc/astrald/node/link"

type EventLinked struct {
	Peer *Peer
	Link *link.Link
}

type EventUnlinked struct {
	Peer *Peer
}

type EventLinkEstablished struct {
	Link *link.Link
}

type EventLinkClosed struct {
	Link *link.Link
}
