package scheduler

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/scheduler"
)

// Schedule enqueues an action for processing by launching a goroutine that
// waits for dependencies, runs the action, and then releases resources.
// It is safe for concurrent use.
func (mod *Module) Schedule(ctx *astral.Context, a scheduler.Action, deps ...scheduler.Waitable) *scheduler.ScheduledAction {
	if a == nil {
		return nil
	}

	ctx, cancel := mod.ctx.WithCancel()
	scheduled := scheduler.NewScheduledAction(a, cancel)

	// If module is shutting down, drop scheduling to avoid starting new work.
	if mod.ctx != nil {
		select {
		case <-mod.ctx.Done():
			mod.log.Log("drop %T: module shutting down", a)
			return &scheduled
		default:
		}
	}

	mod.wg.Add(1)
	go func() {
		defer mod.wg.Done()
		defer scheduled.Close()

		// FIXME: wait for deps to be ready

		err := a.Run(ctx)
		if err != nil {
			mod.log.Errorv(1, "failed to run action %v: %v", a.String(), err)
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
