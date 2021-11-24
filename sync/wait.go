package sync

import "context"

func Wait(ctx context.Context, waiters ...Waiter) error {
	waitCtx, done := context.WithCancel(ctx)
	defer done()

	for _, waiter := range waiters {
		go func() {
			select {
			case <-waiter.Wait():
				waitCtx.Done()
			case <-waitCtx.Done():
				return
			}
		}()
	}

	select {
	case <-waitCtx.Done():
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
