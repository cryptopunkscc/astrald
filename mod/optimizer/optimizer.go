package optimizer

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/infra"
	_log "github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/peers"
	"time"
)

const optimizeDuration = 60 * time.Second
const concurrency = 8

type Module struct {
	node *node.Node
}

var log = _log.Tag(ModuleName)

func (mod *Module) Run(ctx context.Context) error {
	for event := range mod.node.Subscribe(ctx) {
		event := event
		if event, ok := event.(peers.EventPeerLinked); ok {
			nodeId := event.Peer.Identity()
			go func() {
				log.Log("optimizing %s", nodeId)
				if err := mod.Optimize(ctx, event.Peer); err != nil {
					log.Error("optimize %s: %s", nodeId, err)
				} else {
					log.Log("done optimizing %s", nodeId)
				}
			}()
		}
	}

	return nil
}

func (mod *Module) Optimize(parent context.Context, peer *peers.Peer) error {
	ctx, cancel := context.WithTimeout(parent, optimizeDuration)

	// optimize until peer gets unlinked or optimization period ends
	go func() {
		select {
		case <-peer.Wait():
		case <-parent.Done():
		}
		cancel()
	}()

	// use FilterDialer to dial only addresses with potentially better quality score
	filterDialer := NewFilterDialer(mod.node.Infra, func(addr infra.Addr) error {
		sa := scoreAddr(addr)
		sp := scorePeer(peer)

		if sa <= sp {
			return errors.New("score too low")
		}
		return nil
	})

	retryDialer := NewRetryDialer(filterDialer, concurrency)

	go func() {
		list, err := mod.node.Tracker.AddrByIdentity(peer.Identity())
		if err == nil {
			for _, i := range list {
				retryDialer.Add(i)
			}
		}
		for addr := range mod.node.Tracker.Watch(ctx, peer.Identity()) {
			retryDialer.Add(addr)
		}
	}()

	for authConn := range peers.NewConcurrentHandshake(
		mod.node.Identity(),
		peer.Identity(),
		concurrency,
	).Outbound(
		ctx,
		retryDialer.Dial(ctx),
	) {
		mod.node.Peers.AddLink(link.NewFromConn(authConn))
	}

	return nil
}
