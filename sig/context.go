package sig

import (
	"context"
)

// Context executes the function fn. If the function finishes before context ends, Context returns nil.
// If the context ends before the function finishes, Context does not wait for the function and returns ctx.Err().
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
