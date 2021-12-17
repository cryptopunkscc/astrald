package linker

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"sync"
)

func DialMany(ctx context.Context, dialer infra.Dialer, addrCh <-chan infra.Addr, concurrency int) <-chan infra.Conn {
	outCh := make(chan infra.Conn, concurrency)

	// start dialers
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			for addr := range addrCh {
				conn, err := dialer.Dial(ctx, addr)
				if err != nil {
					continue
				}

				outCh <- conn

				select {
				case <-ctx.Done():
					return
				default:
				}
			}
		}()
	}

	// close connection channel once all dialers are done
	go func() {
		wg.Wait()
		close(outCh)
	}()

	return outCh
}
