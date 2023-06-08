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
}

func (m *Module) Run(ctx context.Context) error {
	return tasks.Group(
		&ProfileService{Module: m},
		&EventHandler{Module: m},
	).Run(ctx)
}
