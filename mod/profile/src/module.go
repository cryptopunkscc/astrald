package profile

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/tasks"
)

const serviceName = ".profile"
const serviceType = "node.profile"

type Module struct {
	node  node.Node
	log   *log.Logger
	ctx   context.Context
	sdp   discovery.Module
	nodes nodes.Module
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	return tasks.Group(
		&ProfileService{Module: mod},
		&EventHandler{Module: mod},
	).Run(ctx)
}
