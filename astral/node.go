package astral

import (
	"github.com/cryptopunkscc/astrald/events"
)

// Node defines the basic interface of an astral node
type Node interface {
	Router
	Identity() *Identity
	Events() *events.Queue
}
