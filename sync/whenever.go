package sync

import "context"

func Whenever(ctx context.Context, waiter Waiter, f func()) {
	if f == nil {
		panic("f is nil")
	}

	go func() {
		for {
			select {
			case <-waiter.Wait():
				f()
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
