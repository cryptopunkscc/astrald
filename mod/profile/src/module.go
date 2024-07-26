package profile

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/tasks"
)

const serviceName = ".profile"

type Module struct {
	*routers.PathRouter
	node    astral.Node
	log     *log.Logger
	ctx     context.Context
	nodes   nodes.Module
	exonet  exonet.Module
	dir     dir.Module
	objects objects.Module
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	return tasks.Group(
		&ProfileService{Module: mod},
	).Run(ctx)
}
