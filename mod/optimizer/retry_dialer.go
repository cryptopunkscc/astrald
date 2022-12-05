package optimizer

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/infra"
	"sync"
	"time"
)

type RetryDialer struct {
	dialer      infra.Dialer
	queue       chan infra.Addr
	concurrency int
}

const queueSize = 64
const retryDelay = 5 * time.Second

func NewRetryDialer(dialer infra.Dialer, concurrency int) *RetryDialer {
	return &RetryDialer{
		dialer:      dialer,
		concurrency: concurrency,
		queue:       make(chan infra.Addr, queueSize),
	}
}

func (d *RetryDialer) Add(addr infra.Addr) error {
	select {
	case d.queue <- addr:
		return nil
	default:
		return errors.New("queue overflow")
	}
}

func (d *RetryDialer) Dial(ctx context.Context) <-chan infra.Conn {
	var out = make(chan infra.Conn)
	var wg sync.WaitGroup

	wg.Add(d.concurrency)
	for i := 0; i < d.concurrency; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case addr, ok := <-d.queue:
					if !ok {
						return
					}

					conn, err := d.dialer.Dial(ctx, addr)
					if err == nil {
						out <- conn
					} else {
						d.retry(ctx, addr)
					}
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func (d *RetryDialer) retry(ctx context.Context, addr infra.Addr) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(retryDelay):
		d.Add(addr)
	}
}
