package infra

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/net"
	"sync"
)

var _ Listener = &CoreInfra{}

func (i *CoreInfra) Listen(ctx context.Context) (<-chan net.Conn, error) {
	if len(i.networkDrivers) == 0 {
		return nil, errors.New("no drivers available")
	}

	var output = make(chan net.Conn)
	var wg = sync.WaitGroup{}

	for _, network := range i.Drivers() {
		listener, ok := network.(Listener)
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
