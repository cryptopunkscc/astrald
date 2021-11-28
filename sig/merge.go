package sig

import "context"

// Merge writes to the returned Signal whenever any of the waiters receive a Signal.
// Ignores state-based waiters like Gate.
func Merge(ctx context.Context, waiters ...Waiter) Signal {
	sig := New()

	for _, waiter := range waiters {
		waiter := waiter
		go func() {
			for {
				select {
				case _, ok := <-waiter.Wait():
					// Handle Queue
					if q, ok := waiter.(*Queue); ok {
						if q.Next() == nil {
							return
						}
						waiter = q.Next()
						select {
						case sig <- struct{}{}:
						case <-ctx.Done():
							return
						}
						continue
					}

					if !ok {
						return
					}

					select {
					case sig <- struct{}{}:
					case <-ctx.Done():
						return
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	return sig
}
