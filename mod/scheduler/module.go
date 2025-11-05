package scheduler

import "github.com/cryptopunkscc/astrald/astral"

const ModuleName = "scheduler"

// Module is the public interface other modules depend on.
type Module interface {
	Schedule(ctx *astral.Context, action Action) ScheduledAction
}

type ScheduledAction struct {
	Waitable
	Action Action
}
