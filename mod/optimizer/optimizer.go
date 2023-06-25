package optimizer

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/network"
	"github.com/cryptopunkscc/astrald/node/tracker"
	"time"
)

const optimizeDuration = 60 * time.Second
const concurrency = 8

type Module struct {
	node node.Node
	log  *log.Logger
}

func (mod *Module) Run(ctx context.Context) error {
	return events.Handle(ctx, mod.node.Events(), func(ctx context.Context, event network.EventPeerLinked) error {
		go func() {
			nodeId := event.Peer.Identity()
			mod.log.Log("optimizing %s", nodeId)
			if err := mod.Optimize(ctx, event.Peer); err != nil {
				mod.log.Error("optimize %s: %s", nodeId, err)
			} else {
				mod.log.Log("done optimizing %s", nodeId)
			}
		}()
		return nil
	})
}

func (mod *Module) Optimize(parent context.Context, peer *network.Peer) error {
	ctx, cancel := context.WithTimeout(parent, optimizeDuration)

	// optimize until peer gets unlinked or optimization period ends
	go func() {
		select {
		case <-peer.Done():
		case <-parent.Done():
		}
		cancel()
	}()

	// use FilterDialer to dial only addresses with potentially better quality score
	filterDialer := NewFilterDialer(mod.node.Infra(), func(addr net.Endpoint) error {
		sa := scoreAddr(addr)
		sp := scorePeer(peer)

		if sa <= sp {
			return errors.New("score too low")
		}
		return nil
	})

	retryDialer := NewRetryDialer(filterDialer, concurrency)

	go func() {
		list, err := mod.node.Tracker().EndpointsByIdentity(peer.Identity())
		if err == nil {
			for _, i := range list {
				retryDialer.Add(i)
			}
		}

		events.Handle(ctx, mod.node.Events(), func(ctx context.Context, e tracker.EventNewEndpoint) error {
			if e.Identity.IsEqual(peer.Identity()) {
				retryDialer.Add(e.Endpoint)
			}
			return nil
		})
	}()

	for authConn := range network.NewConcurrentHandshake(
		mod.node.Identity(),
		peer.Identity(),
		concurrency,
	).Outbound(
		ctx,
		retryDialer.Dial(ctx),
	) {
		mod.node.Network().AddLink(link.NewCoreLink(authConn))
	}

	return nil
}
