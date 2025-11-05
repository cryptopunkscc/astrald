package scheduler

import "github.com/cryptopunkscc/astrald/astral"

const ModuleName = "scheduler"

// Module is the public interface other modules depend on.
type Module interface {
	Schedule(ctx *astral.Context, action Action) *ScheduledAction
}

type ScheduledAction struct {
	Waitable
	ScheduledAt astral.Time
	Action      Action
	Done        chan struct{}
}

func (h ScheduledAction) Wait() <-chan struct{} {
	return h.Done
}

func NewScheduledAction(action Action) ScheduledAction {
	return ScheduledAction{
		Action:      action,
		ScheduledAt: astral.Now(),
		Done:        make(chan struct{}),
	}
}
