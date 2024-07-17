package profile

import (
	"context"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	node2 "github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/tasks"
)

const serviceName = ".profile"
const serviceType = "node.profile"

type Module struct {
	*routers.PathRouter
	node   node2.Node
	log    *log.Logger
	ctx    context.Context
	sdp    discovery.Module
	nodes  nodes.Module
	exonet exonet.Module
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	return tasks.Group(
		&ProfileService{Module: mod},
		&EventHandler{Module: mod},
	).Run(ctx)
}
