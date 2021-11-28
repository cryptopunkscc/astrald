package sig

import (
	"context"
)

func Whenever(ctx context.Context, waiter Waiter, fn func()) {
	if fn == nil {
		panic("f is nil")
	}

	go func() {
		waiter := waiter
		for {
			select {
			case <-waiter.Wait():
				// handle Queue
				if q, ok := waiter.(*Queue); ok {
					// If it's EOF, stop here
					if q.Next() == nil {
						return
					}
					waiter = q.Next()
				}

				fn()

				select {
				case <-ctx.Done():
					return
				default:
				}

			case <-ctx.Done():
				return
			}
		}
	}()
}
