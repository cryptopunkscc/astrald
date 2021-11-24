package sync

import "context"

func On(ctx context.Context, waiter Waiter, f func()) {
	if f == nil {
		panic("f is nil")
	}

	go func() {
		select {
		case <-waiter.Wait():
			f()
		case <-ctx.Done():
		}
	}()
}
