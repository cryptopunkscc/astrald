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

func (mod *Module) Prepare(ctx context.Context) error {
	// inject admin command
	if adm, err := admin.Load(mod.node); err == nil {
		adm.AddCommand(ModuleName, NewAdmin(mod))
	}

	return nil
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	tasks.Group(mod.server).Run(ctx)

	<-ctx.Done()

	return nil
}
