package astral

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"log"
	"sync"
)

func (astral *Astral) Listen(ctx context.Context) (<-chan infra.Conn, <-chan error) {
	if astral.networks == nil {
		return nil, nil
	}

	output, errCh := make(chan infra.Conn), make(chan error, 1)
	wg := sync.WaitGroup{}

	for _, network := range astral.networks {
		wg.Add(1)
		go func(network infra.Network) {
			defer wg.Done()
			accept, netErrCh := network.Listen(ctx)
			for conn := range accept {
				output <- conn
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
