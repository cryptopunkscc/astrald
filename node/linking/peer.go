package linking

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/node/contacts"
	"github.com/cryptopunkscc/astrald/node/peer"
	"sync"
)

type PeerOptimizer struct {
	localID  id.Identity
	remoteID id.Identity

	contacts *contacts.Manager
	peers    *peer.Manager
	dialer   infra.Dialer

	optimizers map[string]*NetworkOptimizer
	mu         sync.Mutex
	cancel     context.CancelFunc
}

func NewPeerOptimizer(
	localID id.Identity,
	remoteID id.Identity,
	contacts *contacts.Manager,
	peers *peer.Manager,
	dialer infra.Dialer,
	linkHandler LinkHandlerFunc,
) *PeerOptimizer {
	peerOpt := &PeerOptimizer{
		localID:    localID,
		remoteID:   remoteID,
		contacts:   contacts,
		peers:      peers,
		dialer:     dialer,
		optimizers: make(map[string]*NetworkOptimizer),
	}

	for _, net := range netPriorities {
		peerOpt.optimizers[net] = NewNetworkOptimizer(localID, remoteID, net, contacts, peers, dialer, linkHandler)
	}

	return peerOpt
}

func (o *PeerOptimizer) Start() {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.cancel != nil {
		return
	}

	var ctx context.Context
	ctx, o.cancel = context.WithCancel(context.Background())
	go o.optimize(ctx)
}

func (o *PeerOptimizer) Stop() {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.cancel == nil {
		return
	}
	o.cancel()
	o.cancel = nil
}

func (o *PeerOptimizer) optimize(ctx context.Context) {
	var peer = o.peers.Hold(ctx, o.remoteID)

	best := len(netPriorities)
	for link := range peer.Links() {
		prio := netPriorities.Priority(link.Network())
		if prio < best {
			best = prio
		}
	}

	for i := 0; i < best; i++ {
		o.optimizers[netPriorities[i]].Start()
	}

	links := peer.SubscribeLinks(ctx.Done(), true)

F:
	for {
		select {
		case link := <-links:
			if link == nil {
				break
			}
			prio := netPriorities.Priority(link.Network())
			for i := prio; i < len(netPriorities); i++ {
				o.optimizers[netPriorities[i]].Stop()
			}
		case <-ctx.Done():
			break F
		}
	}

	for _, o := range o.optimizers {
		o.Stop()
	}
}
