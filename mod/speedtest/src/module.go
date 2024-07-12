package speedtest

import (
	"context"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	node2 "github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/tasks"
)

type Module struct {
	node node2.Node
	log  *log.Logger
	ctx  context.Context
}

func (mod *Module) Prepare(ctx context.Context) error {
	// inject admin command
	if adm, err := core.Load[admin.Module](mod.node, admin.ModuleName); err == nil {
		adm.AddCommand(ModuleName, NewAdmin(mod))
	}

	return nil
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	return tasks.Group(&Service{Module: mod}).Run(ctx)
}
