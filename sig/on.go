package sig

import "context"

// OnCtx waits for the waiter to signal or close and invokes fn once.
func OnCtx(ctx context.Context, waiter Signal, fn func()) {
	if fn == nil {
		panic("f is nil")
	}

	go func() {
		select {
		case <-waiter.Done():
			fn()
		case <-ctx.Done():
		}
	}()
}

func On(waiter Signal, fn func()) {
	OnCtx(context.Background(), waiter, fn)
}
