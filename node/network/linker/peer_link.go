package linker

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/network/graph"
	_peer "github.com/cryptopunkscc/astrald/node/network/peer"
	async "github.com/cryptopunkscc/astrald/sync"
	"sync"
	"time"
)

// retryDelay is the duration to wait after an attempt to link over a network. The delay prevents spam.
const retryDelay = 5 * time.Second

// SustainPeerLink ensures there's a link over every network with the peer until context is done.
func SustainPeerLink(ctx context.Context, localID id.Identity, peer *_peer.Peer, resolver graph.Resolver) <-chan *link.Link {
	var outCh = make(chan *link.Link)

	var wg sync.WaitGroup

	for _, network := range astral.NetworkNames() {
		go func(network string) {
			linker := &ConcurrentLinker{
				localID:  localID,
				remoteID: peer.Identity(),
				resolver: graph.FilterNetwork(resolver, network),
			}

			async.Whenever(ctx, _peer.NetworkUnlinkedGate(ctx, peer, network), func() {
				wg.Add(1)
				defer wg.Done()

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
		wg.Wait()
		close(outCh)
	}()

	return outCh
}