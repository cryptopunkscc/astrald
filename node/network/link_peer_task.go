package network

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/link"
)

var ErrNodeUnreachable = errors.New("node unreachable")

// LinkOptions stores options for tasks that create new links with other nodes
type LinkOptions struct {
	// EndpointFilter is a function called by the linker for every address. If it returns false, the address will not
	// be used by the linker.
	EndpointFilter func(addr net.Endpoint) bool
}

// LinkPeerTask represents a task that tries to establish a new link with a node
type LinkPeerTask struct {
	RemoteID id.Identity
	Network  *CoreNetwork
	options  LinkOptions
}

func (task *LinkPeerTask) Run(ctx context.Context) (*link.Link, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Fetch addresses for the remote identity
	endpoints, err := task.Network.tracker.FindAll(task.RemoteID)
	if err != nil {
		return nil, err
	}
	if len(endpoints) == 0 {
		return nil, errors.New("identity has no addresses")
	}

	// Get a list of supported networks
	var networks = task.Network.infra.Drivers()

	// Populate a channel with addresses
	var ch = make(chan net.Endpoint, len(endpoints))
	for _, e := range endpoints {
		if _, found := networks[e.Network()]; !found {
			continue
		}

		if task.options.EndpointFilter != nil {
			if !task.options.EndpointFilter(e) {
				continue
			}
		}
		ch <- e
	}
	close(ch)

	authed := NewConcurrentHandshake(
		task.Network.localID,
		task.RemoteID,
		workers,
	).Outbound(
		ctx,
		NewConcurrentDialer(
			task.Network.infra,
			workers,
		).Dial(
			ctx,
			ch,
		),
	)

	defer func() {
		go func() {
			for a := range authed {
				a.Close()
			}
		}()
	}()

	authConn, ok := <-authed
	if !ok {
		return nil, ErrNodeUnreachable
	}

	l := link.New(authConn)
	l.SetPriority(NetworkPriority(l.Network()))
	if err := task.Network.AddLink(l); err != nil {
		l.Close()
		return nil, err
	}

	return l, nil
}
