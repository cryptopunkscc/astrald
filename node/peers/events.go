package peers

import (
	_log "github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/link"
)

var log = _log.Tag("peers")

type EventPeerLinked struct {
	Peer *Peer
	Link *link.Link
}

type EventPeerUnlinked struct {
	Peer *Peer
}
