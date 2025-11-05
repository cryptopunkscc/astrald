package scheduler

import (
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/scheduler"
	"github.com/cryptopunkscc/astrald/resources"
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

	mu sync.Mutex
	wg sync.WaitGroup
}

func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx

	// Block until module context is done, then wait for in-flight actions to finish
	<-ctx.Done()
	mod.wg.Wait()
	return nil
}

func (mod *Module) String() string {
	return scheduler.ModuleName
}
