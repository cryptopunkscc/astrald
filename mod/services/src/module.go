package services

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/services"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

const ModuleName = "services"

type Module struct {
	Deps

	node astral.Node
	log  *log.Logger
	ops  shell.Scope
	db   *DB

	discoverers []services.ServiceDiscoverer
}

var _ services.Module = &Module{}

func (mod *Module) AddServiceDiscoverer(discoverer services.ServiceDiscoverer) {
	mod.discoverers = append(mod.discoverers, discoverer)
}

func (mod *Module) Run(ctx *astral.Context) error {
	<-ctx.Done()
	return nil
}

func (mod *Module) Scope() *shell.Scope {
	return &mod.ops
}

func (mod *Module) String() string {
	return ModuleName
}
