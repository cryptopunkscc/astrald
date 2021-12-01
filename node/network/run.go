package network

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/link"
	"log"
)

func (n *Network) Run(ctx context.Context, localID id.Identity) (<-chan *link.Query, <-chan Event, <-chan error) {
	ctx, cancel := context.WithCancel(ctx)

	reqCh, evCh, errCh := make(chan *link.Query, 1), make(chan Event, 1), make(chan error, 1)

	go func() {
		defer cancel()
		defer close(reqCh)
		defer close(evCh)
		defer close(errCh)

		err := astral.Announce(ctx, n.localID)
		if err != nil {
			log.Println("announce error:", err)
		}

		discoverCh, err := astral.Discover(ctx)
		if err != nil {
			log.Println("discover error:", err)
		}

		// set up link source
		listenCh, listenErrCh := astral.Listen(ctx, localID)

		for {
			select {
			case rawLink := <-listenCh:
				n.newLinks <- link.Wrap(rawLink)

			case link := <-n.newLinks:
				if err := n.onLink(ctx, link, reqCh, evCh); err != nil {
					log.Println("link rejected:", err)
					link.Close()
				}

			case presence := <-discoverCh:
				if err := n.handlePresence(ctx, presence); err != nil {
					log.Println("error handling presence:", err)
				}

			case err := <-listenErrCh:
				errCh <- err
				return

			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			}
		}
	}()

	return reqCh, evCh, errCh
}
