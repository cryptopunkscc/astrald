package nodes

import (
	"github.com/cryptopunkscc/astrald/id"
)

type EventLinked struct {
	NodeID id.Identity
}

type EventUnlinked struct {
	NodeID id.Identity
}
