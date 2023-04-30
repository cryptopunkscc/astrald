package infra

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/infra"
	"sync"
)

var _ infra.Listener = &Infra{}

func (i *Infra) Listen(ctx context.Context) (<-chan infra.Conn, error) {
	if len(i.networks) == 0 {
		return nil, errors.New("no networks available")
	}

	var output = make(chan infra.Conn)
	var wg = sync.WaitGroup{}

	for _, network := range i.Networks() {
		listener, ok := network.(infra.Listener)
		if !ok {
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()

			// start listening
			listenCh, err := listener.Listen(ctx)
			if err != nil {
				return
			}

			// forward connections to the collective output channel
			for conn := range listenCh {
				output <- conn
			}
		}()
	}

	go func() {
		wg.Wait()
		close(output)
	}()

	return output, nil
}
