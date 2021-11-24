package linker

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/auth/id"
	_peer "github.com/cryptopunkscc/astrald/node/network/peer"
	"github.com/cryptopunkscc/astrald/node/network/route"
	"github.com/cryptopunkscc/astrald/sync"
	"time"
)

// retryDelay is the duration to wait after an attempt to link over a network. The delay prevents spam.
const retryDelay = 5 * time.Second

// SustainPeerLink ensures there's a link over every network with the peer until context is done.
func SustainPeerLink(ctx context.Context, localID id.Identity, peer *_peer.Peer, router route.Router) <-chan *link.Link {
	var outCh = make(chan *link.Link)

	for _, network := range astral.NetworkNames() {
		go func(network string) {
			linker := &ConcurrentLinker{
				localID:  localID,
				remoteID: peer.Identity(),
				router:   route.FilterNetwork(router, network),
			}

			sync.Whenever(ctx, _peer.NetworkUnlinkedGate(ctx, peer, network), func() {
				if l := linker.Link(ctx); l != nil {
					outCh <- l
				}

				select {
				case <-time.After(retryDelay):
				case <-ctx.Done():
				}
			})
		}(network)
	}

	// close the output after all linkers quit
	go func() {
		<-ctx.Done()
		close(outCh)
	}()

	return outCh
}
