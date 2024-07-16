package tor

import (
	"context"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/node"
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

	exonet exonet.Module
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	tasks.Group(mod.server).Run(ctx)

	<-ctx.Done()

	return nil
}

func (mod *Module) Resolve(ctx context.Context, identity id.Identity) (endpoints []exonet.Endpoint, err error) {
	if mod.server == nil {
		return
	}
	if mod.server.endpoint == nil || mod.server.endpoint.IsZero() {
		return
	}

	endpoints = append(endpoints, mod.server.endpoint)

	return
}
