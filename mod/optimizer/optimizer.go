package optimizer

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/infra"
	alink "github.com/cryptopunkscc/astrald/link"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/peers"
	"log"
	"time"
)

const ModuleName = "optimizer"
const optimizeDuration = 60 * time.Second
const logTag = "[optimizer]"
const concurrency = 8

type Optimizer struct {
	node *node.Node
}

func (o *Optimizer) Run(ctx context.Context, n *node.Node) error {
	o.node = n

	for event := range n.Subscribe(ctx.Done()) {
		event := event
		if event, ok := event.(peers.EventLinked); ok {
			peerName := n.Contacts.DisplayName(event.Peer.Identity())
			go func() {
				log.Println(logTag, "optimizing", peerName)
				if err := o.Optimize(ctx, event.Peer); err != nil {
					log.Println(logTag, "optimize", peerName, "error:", err)
				} else {
					log.Println(logTag, "done optimizing", peerName)
				}
			}()
		}
	}

	return nil
}

func (o *Optimizer) Optimize(parent context.Context, peer *peers.Peer) error {
	ctx, cancel := context.WithTimeout(parent, optimizeDuration)
	peerName := o.node.Contacts.DisplayName(peer.Identity())

	// optimize until peer gets unlinked or optimization period ends
	go func() {
		select {
		case <-peer.Wait():
		case <-parent.Done():
		}
		cancel()
	}()

	// use FilterDialer to dial only addresses with potentially better quality score
	filterDialer := NewFilterDialer(o.node.Infra, func(addr infra.Addr) error {
		sa := scoreAddr(addr)
		sp := scorePeer(peer)

		if sa <= sp {
			return errors.New("score too low")
		}
		log.Println("[optimizer] dial", peerName, "at", addr.Network(), addr.String(), sa, ">", sp)
		return nil
	})

	retryDialer := NewRetryDialer(filterDialer, concurrency)

	go func() {
		contact := o.node.Contacts.Find(peer.Identity(), true)
		for addr := range contact.Addr(ctx) {
			retryDialer.Add(addr)
		}
	}()

	for authConn := range peers.NewConcurrentHandshake(
		o.node.Identity(),
		peer.Identity(),
		concurrency,
	).Outbound(
		ctx,
		retryDialer.Dial(ctx),
	) {
		o.node.Peers.AddLink(alink.New(authConn))
	}

	return nil
}

func (*Optimizer) String() string {
	return ModuleName
}
