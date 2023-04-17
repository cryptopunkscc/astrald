package network

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"sync"
	"time"
)

type ConcurrentDialer struct {
	dialer      infra.Dialer
	concurrency int
}

func NewConcurrentDialer(dialer infra.Dialer, concurrency int) *ConcurrentDialer {
	return &ConcurrentDialer{
		dialer:      dialer,
		concurrency: concurrency,
	}
}

func (d *ConcurrentDialer) Dial(ctx context.Context, addrs <-chan infra.Addr) <-chan infra.Conn {
	out := make(chan infra.Conn)

	// spawn workers
	var wg sync.WaitGroup
	wg.Add(d.concurrency)
	for i := 0; i < d.concurrency; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return

				case addr, ok := <-addrs:
					// channel closed?
					if !ok {
						return
					}

					conn, err := d.dialer.Dial(ctx, addr)
					if err != nil {
						return
					}
					select {
					case <-ctx.Done():
						return

					case <-time.After(HandshakeTimeout):
						conn.Close()

					case out <- conn:
					}
				}
			}
		}()
	}

	// close output channel once all workers are done
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
