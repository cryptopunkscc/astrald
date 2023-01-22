package optimizer

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/peers"
	"log"
	"time"
)

const optimizeDuration = 60 * time.Second
const logTag = "[optimizer]"
const concurrency = 8

type Optimizer struct {
	node *node.Node
}

func (mod *Optimizer) Run(ctx context.Context) error {
	for event := range mod.node.Subscribe(ctx) {
		event := event
		if event, ok := event.(peers.EventPeerLinked); ok {
			peerName := mod.node.Contacts.DisplayName(event.Peer.Identity())
			go func() {
				log.Println(logTag, "optimizing", peerName)
				if err := mod.Optimize(ctx, event.Peer); err != nil {
					log.Println(logTag, "optimize", peerName, "error:", err)
				} else {
					log.Println(logTag, "done optimizing", peerName)
				}
			}()
		}
	}

	return nil
}

func (mod *Optimizer) Optimize(parent context.Context, peer *peers.Peer) error {
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
