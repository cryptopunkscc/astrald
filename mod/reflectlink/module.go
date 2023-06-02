package reflectlink

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/discovery/proto"
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

func (mod *Module) Discover(ctx context.Context, caller id.Identity, medium string) ([]proto.ServiceEntry, error) {
	if medium == services.SourceNetwork {
		return []proto.ServiceEntry{{
			Name:  serviceName,
			Type:  serviceType,
			Extra: nil,
		}}, nil
	}

	return []proto.ServiceEntry{}, nil
}
