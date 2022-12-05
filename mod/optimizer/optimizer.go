package optimizer

import (
	"context"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/peers"
	"log"
	"time"
)

const ModuleName = "optimizer"
const optimizeDuration = 60 * time.Second
const logTag = "[optimizer]"

type Optimizer struct{}

func (o Optimizer) Run(ctx context.Context, n *node.Node) error {
	for event := range n.Subscribe(ctx.Done()) {
		if event, ok := event.(peers.EventLinked); ok {
			peerName := n.Contacts.DisplayName(event.Peer.Identity())
			if event.Link.Outbound() {
				event := event
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
	}

	return nil
}

func (Optimizer) Optimize(parent context.Context, peer *peers.Peer) error {
	ctx, cancel := context.WithTimeout(parent, optimizeDuration)

	go func() {
		select {
		case <-peer.Wait():
		case <-parent.Done():
		}
		cancel()
	}()

	//TODO:optimizer

	<-ctx.Done()
	return nil
}

func (Optimizer) String() string {
	return ModuleName
}
