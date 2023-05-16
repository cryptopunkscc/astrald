package optimizer

import (
	"context"
	"errors"
	_log "github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/network"
	"github.com/cryptopunkscc/astrald/node/tracker"
	"time"
)

const optimizeDuration = 60 * time.Second
const concurrency = 8

type Module struct {
	node node.Node
}

var log = _log.Tag(ModuleName)

func (mod *Module) Run(ctx context.Context) error {
	return event.Handle(ctx, mod.node.Events(), func(event network.EventPeerLinked) error {
		go func() {
			nodeId := event.Peer.Identity()
			log.Log("optimizing %s", nodeId)
			if err := mod.Optimize(ctx, event.Peer); err != nil {
				log.Error("optimize %s: %s", nodeId, err)
			} else {
				log.Log("done optimizing %s", nodeId)
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
		list, err := mod.node.Tracker().FindAll(peer.Identity())
		if err == nil {
			for _, i := range list {
				retryDialer.Add(i)
			}
		}

		event.Handle(ctx, mod.node.Events(), func(e tracker.EventNewEndpoint) error {
			if e.Identity.IsEqual(peer.Identity()) {
				retryDialer.Add(e.TrackedEndpoint.Endpoint)
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
		l := link.New(authConn)
		l.SetPriority(network.NetworkPriority(l.Network()))
		mod.node.Network().AddLink(l)
	}

	return nil
}
