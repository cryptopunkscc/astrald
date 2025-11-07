package scheduler

import (
	"context"

	"github.com/cryptopunkscc/astrald/astral"
)

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
	cancel      context.CancelFunc
}

func (h ScheduledAction) Wait() <-chan struct{} {
	return h.done
}

// NOTE: maybe Done()? but then it is similiar to context.Done()
func (h ScheduledAction) Close() {
	close(h.done)
}

func (h ScheduledAction) Cancel() {
	h.cancel()
}

func NewScheduledAction(action Action, cancel context.CancelFunc) ScheduledAction {
	return ScheduledAction{
		Action:      action,
		ScheduledAt: astral.Now(),
		done:        make(chan struct{}),
		cancel:      cancel,
	}
}
