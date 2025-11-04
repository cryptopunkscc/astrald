package scheduler

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/scheduler"
)

// runWorker processes scheduled actions sequentially until context is canceled
func (mod *Module) runWorker(ctx *astral.Context) {
	mod.log.Info("worker started")
	for a := range mod.q.Subscribe(ctx) {
		if a == nil {
			continue
		}
		mod.log.Log("start %T", a)
		if err := a.Run(ctx); err != nil {
			mod.log.Error("fail %T: %v", a, err)
		} else {
			mod.log.Log("done %T", a)
		}
	}
	mod.log.Info("worker stopped")
}

// Schedule enqueues an action for processing. It is safe for concurrent use.
func (mod *Module) Schedule(ctx *astral.Context, a scheduler.Action) {
	if a == nil {
		return
	}
	// avoid pushes after shutdown
	if mod.ctx != nil {
		select {
		case <-mod.ctx.Done():
			mod.log.Log("drop %T: module shutting down", a)
			return
		default:
		}
	}

	_ = mod.q.Push(a)
}
