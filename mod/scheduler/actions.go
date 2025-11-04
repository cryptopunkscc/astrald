package scheduler

import "github.com/cryptopunkscc/astrald/astral"

// Action represents a unit of work to be executed by the scheduler.
type Action interface {
	Run(*astral.Context) error
}

type Waitable interface {
	Done() <-chan struct{}
}

// NOTE: just demonstration purposes
type Preparable interface {
	Prepare(ctx *astral.Context) error
}
