package tor

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/tasks"
	"golang.org/x/net/proxy"
)

const defaultListenPort = 1791

type Module struct {
	config Config
	node   node.Node
	nodes  nodes.Module
	assets assets.Assets
	log    *log.Logger
	ctx    context.Context
	proxy  proxy.ContextDialer
	server *Server
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	tasks.Group(mod.server).Run(ctx)

	<-ctx.Done()

	return nil
}
