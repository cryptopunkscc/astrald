package scheduler

import (
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/events"
)

// Action represents a unit of work to be executed by the scheduler.
type Action interface {
	fmt.Stringer
	Run(*astral.Context) error
}

type Doner interface {
	Done() <-chan struct{}
}

type ResourceReleaser interface {
	Release()
}

type EventReceiver interface {
	ReceiveEvent(e *events.Event)
}
