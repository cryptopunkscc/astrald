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

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	// inject admin command
	if adm, _ := mod.node.Modules().Find("admin").(admin.API); adm != nil {
		adm.AddCommand("speedtest", NewAdmin(mod))
	}

	return tasks.Group(&Service{Module: mod}).Run(ctx)
}
