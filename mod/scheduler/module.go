package scheduler

import "github.com/cryptopunkscc/astrald/astral"

const ModuleName = "scheduler"

// Module is the public interface other modules depend on.
type Module interface {
	Schedule(ctx *astral.Context, action Action, deps ...Waitable) *ScheduledAction
}

type ScheduledAction struct {
	Waitable
	ScheduledAt astral.Time
	Action      Action
	done        chan struct{}
}

func (h ScheduledAction) Wait() <-chan struct{} {
	return h.done
}

// NOTE: maybe Done()? but then it is similiar to context.Done()
func (h ScheduledAction) Close() {
	close(h.done)
}

func NewScheduledAction(action Action) ScheduledAction {
	return ScheduledAction{
		Action:      action,
		ScheduledAt: astral.Now(),
		done:        make(chan struct{}),
	}
}
