package scheduler

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/scheduler"
)

// runWorker processes scheduled actions sequentially until context is canceled
func (mod *Module) runWorker(ctx *astral.Context) {
	mod.log.Info("worker started")
	for scheduledAction := range mod.q.Subscribe(ctx) {
		mod.log.Log("start %T", scheduledAction)
		if err := scheduledAction.Action.Run(ctx); err != nil {
			mod.log.Error("fail %T: %v", scheduledAction, err)
		} else {
			mod.log.Log("done %T", scheduledAction)
		}
	}
	mod.log.Info("worker stopped")
}

// Schedule enqueues an action for processing. It is safe for concurrent use.
func (mod *Module) Schedule(ctx *astral.Context, a scheduler.Action) *scheduler.ScheduledAction {
	if a == nil {
		return nil
	}

	scheduled := scheduler.NewScheduledAction(a)
	if mod.ctx != nil {
		select {
		case <-mod.ctx.Done():
			mod.log.Log("drop %T: module shutting down", a)
			return &scheduled
		default:
		}
	}

	_ = mod.q.Push(scheduled)
	return &scheduled
}
