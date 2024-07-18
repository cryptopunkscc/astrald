package astral

import (
	"github.com/cryptopunkscc/astrald/events"
	"github.com/cryptopunkscc/astrald/id"
)

// Node defines the basic interface of an astral node
type Node interface {
	Identity() id.Identity
	Router() Router
	Events() *events.Queue
}
