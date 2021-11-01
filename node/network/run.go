package network

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/auth/id"
	"log"
)

func (network *Network) Run(ctx context.Context, localID id.Identity) (<-chan link.Request, <-chan Event, <-chan error) {
	ctx, cancel := context.WithCancel(ctx)

	network.loadState()
	defer network.storeState()

	reqCh, evCh, errCh := make(chan link.Request, 1), make(chan Event, 1), make(chan error, 1)

	go func() {
		defer cancel()
		defer close(reqCh)
		defer close(evCh)
		defer close(errCh)

		err := astral.Announce(ctx, network.identity)
		if err != nil {
			log.Println("announce error:", err)
		}

		discoverCh, err := astral.Discover(ctx)
		if err != nil {
			log.Println("discover error:", err)
		}

		// set up link source
		listenCh, listenErrCh := astral.Listen(ctx, localID)
		linkerCh := network.Linker.Run(ctx)

		linksCh := mergeLinkChans(listenCh, linkerCh)

		for {
			select {
			case link := <-linksCh:
				if err := network.onLink(ctx, link, reqCh, evCh); err != nil {
					link.Close()
				}

			case presence := <-discoverCh:
				network.Router.AddAddr(presence.Identity, presence.Addr)

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
