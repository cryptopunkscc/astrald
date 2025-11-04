package scheduler

import (
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/scheduler"
	"github.com/cryptopunkscc/astrald/resources"
	"github.com/cryptopunkscc/astrald/sig"
)

// Ensure Module struct implements the public scheduler.Module interface
var _ scheduler.Module = (*Module)(nil)

// Module is the concrete implementation of the scheduler module.
type Module struct {
	Deps

	ctx    *astral.Context
	node   astral.Node
	log    *log.Logger
	assets resources.Resources

	q  *sig.Queue[scheduler.Action]
	wg sync.WaitGroup
}

func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx
	// start worker
	mod.wg.Add(1)
	go func() {
		defer mod.wg.Done()
		mod.runWorker(ctx)
	}()

	// Probably we will have somekind mechanisms that will allow to wait for
	// actions before closing node.
	<-ctx.Done()
	// graceful stop: close queue and wait for worker to finish
	if mod.q != nil {
		mod.q.Close()
	}
	mod.wg.Wait()
	return nil
}

func (mod *Module) String() string {
	return scheduler.ModuleName
}

func (mod *Module) WaitableAction(a scheduler.Action) scheduler.Action {
	return NewWaitable(a)
}
