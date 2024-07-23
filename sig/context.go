package sig

import (
	"context"
)

func Context(ctx context.Context, fn func()) error {
	ch := make(chan error, 1)

	go func() {
		fn()
		ch <- nil
	}()

	go func() {
		<-ctx.Done()
		ch <- ctx.Err()
	}()

	return <-ch
}
