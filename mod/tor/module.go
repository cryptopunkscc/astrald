package tor

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin/api"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/tasks"
	"golang.org/x/net/proxy"
)

const defaultListenPort = 1791

type Module struct {
	config Config
	node   node.Node
	assets assets.Store
	log    *log.Logger
	ctx    context.Context
	proxy  proxy.ContextDialer
	server *Server
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	// inject admin command
	if adm, _ := mod.node.Modules().Find("admin").(admin.API); adm != nil {
		adm.AddCommand(ModuleName, NewAdmin(mod))
	}

	tasks.Group(mod.server).Run(ctx)

	<-ctx.Done()

	return nil
}
