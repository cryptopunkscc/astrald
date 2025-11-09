package scheduler

import (
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
)

// Action represents a unit of work to be executed by the scheduler.
type Action interface {
	fmt.Stringer
	Run(*astral.Context) error
}

type Doner interface {
	Done() <-chan struct{}
}

type ResourceHolder interface {
	Doner
	Release()
}
