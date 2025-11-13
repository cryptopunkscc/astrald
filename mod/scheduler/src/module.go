package scheduler

import (
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

	queue sig.Set[scheduler.ScheduledTask]
}

func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx

	<-ctx.Done()
	return nil
}

func (mod *Module) String() string {
	return scheduler.ModuleName
}
