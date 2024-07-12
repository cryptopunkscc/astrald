package reflectlink

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/net"
	node2 "github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/tasks"
)

const serviceName = ".reflect"
const serviceType = "mod.reflectlink"

type Module struct {
	node node2.Node
	log  *log.Logger
	sdp  discovery.Module
}

func (mod *Module) Run(ctx context.Context) error {
	return tasks.Group(
		&Server{Module: mod},
		&Client{Module: mod},
	).Run(ctx)
}

func (mod *Module) DiscoverServices(ctx context.Context, caller id.Identity, origin string) ([]discovery.Service, error) {
	if origin == net.OriginNetwork {
		return []discovery.Service{{
			Identity: mod.node.Identity(),
			Name:     serviceName,
			Type:     serviceType,
			Extra:    nil,
		}}, nil
	}

	return []discovery.Service{}, nil
}
