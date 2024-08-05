package profile

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/tasks"
)

const serviceName = ".profile"

var _ objects.Receiver = &Module{}

type Module struct {
	Deps
	*routers.PathRouter
	node astral.Node
	log  *log.Logger
	ctx  context.Context
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	return tasks.Group(
		&ProfileService{Module: mod},
	).Run(ctx)
}

func (mod *Module) String() string {
	return ModuleName
}
