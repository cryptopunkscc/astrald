package scheduler

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/events"
	"github.com/cryptopunkscc/astrald/mod/objects"
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

	queue sig.Set[scheduler.ScheduledAction]
}

func (mod *Module) ReceiveObject(drop objects.Drop) (err error) {
	switch o := drop.Object().(type) {
	case *events.Event:
		for _, a := range mod.queue.Clone() {
			if a.State() == scheduler.ScheduledActionStateRunning {
				if a, ok := a.Action().(scheduler.EventReceiver); ok {
					a.ReceiveEvent(o)
				}
			}
		}
	}

	return nil
}
func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx

	// Block until module context is done, then wait for in-flight actions to finish
	<-ctx.Done()
	return nil
}

func (mod *Module) String() string {
	return scheduler.ModuleName
}
