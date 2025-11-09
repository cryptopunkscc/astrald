package scheduler

import (
	"context"
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/scheduler"
)

// Schedule enqueues an action for processing by launching a goroutine that
// waits for dependencies, runs the action, and then releases resources.
// It is safe for concurrent use.
func (mod *Module) Schedule(ctx *astral.Context, a scheduler.Action, deps ...scheduler.Doner) scheduler.ScheduledAction {
	if a == nil {
		return nil
	}

	actionCtx, cancel := ctx.WithCancelCause()
	scheduled := NewScheduledAction(a, cancel)

	// If module is shutting down, drop scheduling to avoid starting new work.
	if mod.ctx != nil {
		select {
		case <-mod.ctx.Done():
			mod.log.Log("drop %v: module shutting down", a.String())
			return &scheduled
		default:
		}
	}

	err := mod.queue.Add(&scheduled)
	if err != nil {
		mod.log.Errorv(1, "failed to add action %v to queue: %v", a.String(), err)
		return &scheduled
	}

	go func() {
		defer scheduled.close()
		defer mod.queue.Remove(&scheduled)

		// FIXME: wait for deps to be ready

		for _, d := range deps {
			select {
			case <-actionCtx.Done():
				scheduled.err = ctx.Err()
				return
			case <-d.Done():
			}
		}

		select {
		case <-actionCtx.Done():
			scheduled.err = ctx.Err()
			return
		default:
			break
		}

		scheduled.err = a.Run(actionCtx)
		if scheduled.err != nil {
			mod.log.Errorv(1, "failed to run action %v: %v", a.String(), scheduled.err)
		}

		// After execution, release resources if deps are ResourceHolders
		for _, d := range deps {
			if rh, ok := d.(scheduler.ResourceHolder); ok && rh != nil {
				rh.Release()
			}
		}
	}()

	return &scheduled
}

type ScheduledAction struct {
	scheduledAt astral.Time
	action      scheduler.Action
	done        chan struct{}
	cancel      context.CancelCauseFunc
	cancelOnce  *sync.Once
	err         error
}

func (h ScheduledAction) Err() error {
	return h.err
}

func (h ScheduledAction) CancelWithError(err error) {
	h.err = err
	h.cancel(err)
	return
}

func (h ScheduledAction) Done() <-chan struct{} {
	return h.done
}

// called externally
func (h ScheduledAction) Cancel() {
	h.cancel(context.Canceled)
	h.close()
	return
}

// called by scheduler
func (h ScheduledAction) close() {
	h.cancelOnce.Do(func() {
		close(h.done)
	})

	return
}

func (h ScheduledAction) Action() scheduler.Action {
	return h.action
}

func (h ScheduledAction) ScheduledAt() astral.Time {
	return h.scheduledAt
}

func NewScheduledAction(action scheduler.Action,
	cancelCauseFunc context.CancelCauseFunc) ScheduledAction {
	return ScheduledAction{
		action:      action,
		scheduledAt: astral.Now(),
		done:        make(chan struct{}),
		cancel:      cancelCauseFunc,
		cancelOnce:  &sync.Once{},
	}
}
