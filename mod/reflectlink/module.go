package reflectlink

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/sdp/api"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/tasks"
)

const serviceName = "net.reflectlink"
const serviceType = "net.reflectlink"

type Module struct {
	node node.Node
	log  *log.Logger
	sdp  sdp.API
}

func (mod *Module) Prepare(ctx context.Context) error {
	mod.sdp, _ = sdp.Load(mod.node)

	return nil
}

func (mod *Module) Run(ctx context.Context) error {
	return tasks.Group(
		&Server{Module: mod},
		&Client{Module: mod},
	).Run(ctx)
}

func (mod *Module) Discover(ctx context.Context, caller id.Identity, origin string) ([]sdp.ServiceEntry, error) {
	if origin == net.OriginNetwork {
		return []sdp.ServiceEntry{{
			Identity: mod.node.Identity(),
			Name:     serviceName,
			Type:     serviceType,
			Extra:    nil,
		}}, nil
	}

	return []sdp.ServiceEntry{}, nil
}
