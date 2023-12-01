package tcp

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin/api"
	. "github.com/cryptopunkscc/astrald/mod/tcp/api"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/tasks"
)

var _ API = &Module{}

type Module struct {
	config          Config
	node            node.Node
	log             *log.Logger
	ctx             context.Context
	publicEndpoints []Endpoint
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	// inject admin command
	if adm, _ := mod.node.Modules().Find("admin").(admin.API); adm != nil {
		adm.AddCommand(ModuleName, NewAdmin(mod))
	}

	tasks.Group(NewServer(mod)).Run(ctx)

	<-ctx.Done()

	return nil
}

func (mod *Module) ListenPort() int {
	return mod.config.ListenPort
}
