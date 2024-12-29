package tor

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/tasks"
	"golang.org/x/net/proxy"
)

const defaultListenPort = 1791

type Deps struct {
	Admin  admin.Module
	Nodes  nodes.Module
	Exonet exonet.Module
}

type Module struct {
	Deps
	config Config
	node   astral.Node
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

func (mod *Module) ResolveEndpoints(ctx context.Context, identity *astral.Identity) (endpoints []exonet.Endpoint, err error) {
	if mod.server == nil {
		return
	}
	if mod.server.endpoint == nil || mod.server.endpoint.IsZero() {
		return
	}

	endpoints = append(endpoints, mod.server.endpoint)

	return
}
