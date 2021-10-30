package astral

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/auth"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"log"
	"sync"
)

func (astral *Astral) Listen(ctx context.Context, localID id.Identity) (<-chan *link.Link, <-chan error) {
	if astral.networks == nil {
		return nil, nil
	}

	output, errCh := make(chan *link.Link), make(chan error, 1)
	wg := sync.WaitGroup{}

	for _, network := range astral.networks {
		wg.Add(1)
		go func(network infra.Network) {
			defer wg.Done()
			accept, netErrCh := network.Listen(ctx)
			for conn := range accept {
				authConn, err := auth.HandshakeInbound(ctx, conn, localID)
				if err != nil {
					conn.Close()
					continue
				}

				output <- link.New(authConn)
			}

			err := <-netErrCh
			if err != nil {
				log.Println(network.Name(), "listen error:", err)
			}
		}(network)
	}

	go func() {
		wg.Wait()
		close(output)
		close(errCh)
	}()

	return output, errCh
}
