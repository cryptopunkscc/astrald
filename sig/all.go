package sig

import (
	"context"
	"sync"
	"sync/atomic"
)

// All waits for all waiters and writes to the returned Signal once.
func All(ctx context.Context, waiters ...Waiter) Signal {
	sig := New()

	go func() {
		var l = len(waiters)
		var n uint32
		var wg sync.WaitGroup

		wg.Add(l)

		for _, waiter := range waiters {
			waiter := waiter
			go func() {
				defer wg.Done()
				select {
				case _, ok := <-waiter.Wait():
					if ok {
						atomic.AddUint32(&n, 1)
						return
					}

					// check for Queue EOF
					if q, ok := waiter.(*Queue); ok {
						if q.Next() == nil {
							return
						}
					}

					atomic.AddUint32(&n, 1)

				case <-ctx.Done():
				}
			}()
		}

		wg.Wait()
		if n == uint32(l) {
			select {
			case sig <- struct{}{}:
			case <-ctx.Done():
				return
			}
		}
	}()

	return sig
}
