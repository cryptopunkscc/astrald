package network

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	nodeinfra "github.com/cryptopunkscc/astrald/node/infra"
	"github.com/cryptopunkscc/astrald/node/link"
)

var ErrNodeUnreachable = errors.New("node unreachable")

// LinkOptions stores options for tasks that create new links with other nodes
type LinkOptions struct {
	// AddrFilter is a function called by the linker for every address. If it returns false, the address will not
	// be used by the linker.
	AddrFilter func(addr infra.Addr) bool
}

// LinkPeerTask represents a task that tries to establish a new link with a node
type LinkPeerTask struct {
	RemoteID id.Identity
	Network  *Network
	options  LinkOptions
}

func (task *LinkPeerTask) Run(ctx context.Context) (*link.Link, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Fetch addresses for the remote identity
	addrs, err := task.Network.tracker.AddrByIdentity(task.RemoteID)
	if err != nil {
		return nil, err
	}
	if len(addrs) == 0 {
		return nil, errors.New("identity has no addresses")
	}

	// Populate a channel with addresses
	addrCh := make(chan infra.Addr, len(addrs))
	for _, a := range addrs {
		for n := range task.Network.infra.Networks() {
			if a.Network() == n {
				goto supported
			}
		}
		continue

	supported:
		if task.options.AddrFilter != nil {
			if !task.options.AddrFilter(a) {
				continue
			}
		}
		addrCh <- a
	}
	close(addrCh)

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
			addrCh,
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
	l.SetPriority(nodeinfra.NetworkPriority(l.Network()))
	if err := task.Network.AddLink(l); err != nil {
		log.Errorv(1, "LinkPeerTask: error adding link to network: %s", err)
		l.Close()
		return nil, err
	}

	return l, nil
}
