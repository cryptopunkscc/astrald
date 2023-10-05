package gateway

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/tasks"
)

type Module struct {
	node node.Node
	log  *log.Logger
	ctx  context.Context
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	return tasks.Group(
		&GatewayService{Module: mod},
		&ConnectService{Module: mod},
	).Run(ctx)
}
