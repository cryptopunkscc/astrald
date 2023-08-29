package speedtest

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/modules"
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
	if adm, err := modules.Find[*admin.Module](mod.node.Modules()); err == nil {
		adm.AddCommand("speedtest", NewAdmin(mod))
	}

	return tasks.Group(&Service{Module: mod}).Run(ctx)
}
