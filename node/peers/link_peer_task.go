package peers

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	ainfra "github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/node/link"
)

var ErrNodeUnreachable = errors.New("node unreachable")

// LinkOptions stores options for tasks that create new links with other nodes
type LinkOptions struct {
	// AddrFilter is a function called by the linker for every address. If it returns false, the address will not
	// be used by the linker.
	AddrFilter func(addr ainfra.Addr) bool
}

// LinkPeerTask represents a task that tries to establish a new link with a node
type LinkPeerTask struct {
	RemoteID id.Identity
	Peers    *Manager
	options  LinkOptions
}

func (task *LinkPeerTask) Run(ctx context.Context) (*link.Link, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Fetch addresses for the remote identity
	addrs, err := task.Peers.tracker.AddrByIdentity(task.RemoteID)
	if err != nil {
		return nil, err
	}
	if len(addrs) == 0 {
		return nil, errors.New("identity has no addresses")
	}

	// Populate a channel with addresses
	addrCh := make(chan ainfra.Addr, len(addrs))
	for _, a := range addrs {
		if task.options.AddrFilter != nil {
			if !task.options.AddrFilter(a) {
				continue
			}
		}
		addrCh <- a
	}
	close(addrCh)

	authed := NewConcurrentHandshake(
		task.Peers.localID,
		task.RemoteID,
		workers,
	).Outbound(
		ctx,
		NewConcurrentDialer(
			task.Peers.infra,
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

	l := link.NewFromConn(authConn)
	if err := task.Peers.AddLink(l); err != nil {
		l.Close()
		return nil, err
	}

	return l, nil
}
