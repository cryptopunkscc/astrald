package profile

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/tasks"
)

const serviceName = "sys.profile"
const serviceType = "sys.profile"

type Module struct {
	node node.Node
	log  *log.Logger
	ctx  context.Context
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx
	return tasks.Group(
		&ProfileService{Module: mod},
		&EventHandler{Module: mod},
	).Run(ctx)
}
