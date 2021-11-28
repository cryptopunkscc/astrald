package sig

import "context"

// On waits for the waiter to signal or close and invokes fn once.
func On(ctx context.Context, waiter Waiter, fn func()) {
	if fn == nil {
		panic("f is nil")
	}

	go func() {
		select {
		case <-waiter.Wait():
			fn()
		case <-ctx.Done():
		}
	}()
}
