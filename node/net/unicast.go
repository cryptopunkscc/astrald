package net

import (
	"context"
	"sync"
)

func Dial(ctx context.Context, addr Addr) (Conn, error) {
	driver := unicastNets[addr.Network()]

	if driver == nil {
		return nil, ErrUnsupportedNetwork
	}

	return driver.Dial(ctx, addr)
}

func Listen(ctx context.Context) <-chan Conn {
	connCh := make(chan Conn)

	wg := sync.WaitGroup{}

	for _, drv := range unicastNets {
		ch, err := drv.Listen(ctx)
		if err != nil {
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range ch {
				connCh <- i
			}
		}()
	}

	go func() {
		wg.Wait()
		close(connCh)
	}()

	return connCh
}
