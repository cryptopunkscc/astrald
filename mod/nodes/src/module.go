package nodes

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/resources"
	"github.com/cryptopunkscc/astrald/tasks"
)

var _ nodes.Module = &Module{}

type Module struct {
	config Config
	node   node.Node
	log    *log.Logger
	assets resources.Resources

	dir dir.Module
}

func (mod *Module) Run(ctx context.Context) error {
	return tasks.Group(
		&Service{Module: mod},
	).Run(ctx)
}

func (mod *Module) Resolve(ctx context.Context, identity id.Identity, opts *nodes.ResolveOpts) ([]net.Endpoint, error) {
	return mod.node.Tracker().EndpointsByIdentity(identity)
}

func (mod *Module) AddEndpoint(nodeID id.Identity, endpoint net.Endpoint) error {
	return mod.node.Tracker().AddEndpoint(nodeID, endpoint)
}

func (mod *Module) RemoveEndpoint(nodeID id.Identity, endpoint net.Endpoint) error {
	panic("implement me")
}

func (mod *Module) Forget(nodeID id.Identity) error {
	return mod.node.Tracker().Clear(nodeID)
}
