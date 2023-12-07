package speedtest

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin/api"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/tasks"
)

type Module struct {
	node node.Node
	log  *log.Logger
	ctx  context.Context
}

func (mod *Module) Prepare(ctx context.Context) error {
	// inject admin command
	if adm, err := admin.Load(mod.node); err == nil {
		adm.AddCommand(ModuleName, NewAdmin(mod))
	}

	return nil
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	return tasks.Group(&Service{Module: mod}).Run(ctx)
}
