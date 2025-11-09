package scheduler

import (
	"github.com/cryptopunkscc/astrald/astral"
)

const ModuleName = "scheduler"

// Module is the public interface other modules depend on.
type Module interface {
	Schedule(ctx *astral.Context, action Action, deps ...Doner) ScheduledAction
}

type ScheduledAction interface {
	Doner
	Action() Action
	ScheduledAt() astral.Time
	CancelWithError(error)
	Cancel()
	Err() error
}
