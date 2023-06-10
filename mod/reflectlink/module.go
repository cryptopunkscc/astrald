package reflectlink

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/services"
	"github.com/cryptopunkscc/astrald/tasks"
)

const serviceName = "net.reflectlink"
const serviceType = "net.reflectlink"

type Module struct {
	node node.Node
	log  *log.Logger
}

func (mod *Module) Run(ctx context.Context) error {
	return tasks.Group(
		&Server{Module: mod},
		&Client{Module: mod},
	).Run(ctx)
}

func (mod *Module) Discover(ctx context.Context, caller id.Identity, origin string) ([]discovery.ServiceEntry, error) {
	if origin == services.OriginNetwork {
		return []discovery.ServiceEntry{{
			Identity: mod.node.Identity(),
			Name:     serviceName,
			Type:     serviceType,
			Extra:    nil,
		}}, nil
	}

	return []discovery.ServiceEntry{}, nil
}
