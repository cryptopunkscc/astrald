package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/ip"
	"github.com/cryptopunkscc/astrald/mod/nat"

	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/resources"
)

// Ensure Module struct implements the public nat.Module interface
var _ nat.Module = &Module{}

// Deps are injected by the core injector.
type Deps struct {
	Dir     dir.Module
	Objects objects.Module
	IP      ip.Module
}

// Module is the concrete implementation of the NAT module.
type Module struct {
	Deps

	ctx    *astral.Context
	node   astral.Node
	log    *log.Logger
	assets resources.Resources

	ops shell.Scope
}

// Run blocks until the context is done.
func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx.IncludeZone(astral.ZoneNetwork)

	<-ctx.Done()
	return nil
}
