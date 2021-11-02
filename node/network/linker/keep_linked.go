package linker

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/node/network/peer"
	"github.com/cryptopunkscc/astrald/node/network/route"
	"sync"
	"time"
)

const concurrentDialers = 4
const linkCooldown = 10 * time.Second
const postLinkDelay = 3 * time.Second
const activeLinkPeriod = 5 * time.Minute

func KeepLinked(ctx context.Context, localID id.Identity, peer *peer.Peer, router route.Router) <-chan *link.Link {
	outCh := make(chan *link.Link)

	var wg sync.WaitGroup
	for _, network := range astral.NetworkNames() {
		network := network
		wg.Add(1)
		go func() {
			defer wg.Done()
			for link := range KeepLinkedViaNetwork(ctx, localID, peer, router, network) {
				outCh <- link
			}
		}()
	}

	go func() {
		wg.Wait()
		close(outCh)
	}()

	return outCh
}

func KeepLinkedViaNetwork(ctx context.Context, localID id.Identity, peer *peer.Peer, router route.Router, network string) <-chan *link.Link {
	outCh := make(chan *link.Link)
	go func() {
		defer close(outCh)

		for {
			// make a channel that will be closed when the peer goes offline on the network
			offlineCh := make(chan struct{})
			go func() {
				_ = waitNetworkOffline(ctx, peer, network)
				close(offlineCh)
			}()

			select {
			case <-ctx.Done():
				return
			case <-offlineCh:
			}

			// get current route
			route := router.Route(peer.Identity())

			if route == nil {
				select {
				case <-ctx.Done():
					return
				case <-time.After(linkCooldown):
				}
				continue
			}

			addrCh := make(chan infra.Addr)
			connCh := make(chan infra.Conn)

			// start dialers
			var dialGroup sync.WaitGroup
			dialGroup.Add(concurrentDialers)
			for i := 0; i < concurrentDialers; i++ {
				go func() {
					for addr := range addrCh {
						conn, err := astral.Dial(addr)
						if err == nil {
							connCh <- conn
						}
					}
					dialGroup.Done()
				}()
			}

			// close connection channel once all dialers are done
			go func() {
				dialGroup.Wait()
				close(connCh)
			}()

			// feed addresses to the dialers
			for _, addr := range route.Addresses {
				if addr.Network() == network {
					addrCh <- addr
				}
			}
			close(addrCh)

			var peerLink *link.Link

			// go through conns until link is established
			for conn := range connCh {
				link, err := astral.Link(localID, peer.Identity(), conn)
				if err != nil {
					continue
				}

				peerLink = link
				break
			}

			// remaining connections (if any) are not needed so just close them
			go func() {
				for conn := range connCh {
					conn.Close()
				}
			}()

			if peerLink != nil {
				outCh <- peerLink
				select {
				case <-ctx.Done():
					return
				case <-time.After(postLinkDelay):
				}
				continue
			}

			// cool down after failure to link
			select {
			case <-ctx.Done():
				return
			case <-time.After(linkCooldown):
			}
		}
	}()

	return outCh
}

func waitNetworkOffline(ctx context.Context, peer *peer.Peer, network string) error {
	for {
		// are we already unlinked on this network?
		if !peer.IsLinkedVia(network) {
			return nil
		}

		// if not, wait for all links to close
		var wg sync.WaitGroup
		for link := range peer.Links() {
			if link.Network() == network {
				wg.Add(1)
				go func() {
					select {
					case <-link.WaitClose():
					case <-ctx.Done():
					}
					wg.Done()
				}()
			}
		}
		wg.Wait()

		// check if context is done
		select {
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
