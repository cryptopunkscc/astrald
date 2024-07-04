package nodes

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/nodes/src/muxlink"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/resources"
	"github.com/cryptopunkscc/astrald/tasks"
	"time"
)

const DefaultWorkerCount = 8
const DefaultTimeout = time.Minute

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

func (mod *Module) AcceptLink(ctx context.Context, conn net.Conn) (net.Link, error) {
	l, err := muxlink.Accept(ctx, conn, mod.node.Identity(), mod.node.LocalRouter())
	if err != nil {
		return nil, err
	}

	err = mod.node.Network().AddLink(l)
	if err != nil {
		l.Close()
	}

	return l, err
}

func (mod *Module) InitLink(ctx context.Context, conn net.Conn, remoteID id.Identity) (net.Link, error) {
	l, err := muxlink.Open(ctx, conn, remoteID, mod.node.Identity(), mod.node.LocalRouter())
	if err != nil {
		return nil, err
	}

	err = mod.node.Network().AddLink(l)
	if err != nil {
		l.Close()
	}

	return l, err
}

func (mod *Module) Link(ctx context.Context, remoteIdentity id.Identity, opts nodes.LinkOpts) (net.Link, error) {
	l, err := (&Linker{mod}).LinkOpts(ctx, remoteIdentity, opts)
	if err != nil {
		return nil, err
	}

	err = mod.node.Network().AddLink(l)
	if err != nil {
		l.Close()
	}

	return l, err
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
