package sig

import "context"

// OnCtx waits for the waiter to signal or close and invokes fn once.
func OnCtx(ctx context.Context, waiter Waiter, fn func()) {
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

func On(waiter Waiter, fn func()) {
	OnCtx(context.Background(), waiter, fn)
}
