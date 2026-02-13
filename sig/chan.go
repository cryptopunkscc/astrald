package sig

import (
	"context"
)

// ChanToArray drains ch until it is closed and returns all values in arrival order.
func ChanToArray[T any](ch <-chan T) (arr []T) {
	for i := range ch {
		arr = append(arr, i)
	}
	return
}

// ArrayToChan converts arr into a closed, buffered channel containing all elements.
func ArrayToChan[T any](arr []T) <-chan T {
	var ch = make(chan T, len(arr))
	for _, i := range arr {
		ch <- i
	}
	close(ch)
	return ch
}

func FilterChan[T any](
	in <-chan T,
	predicate func(T) bool,
) <-chan T {
	out := make(chan T)

	go func() {
		defer close(out)
		for v := range in {
			if predicate(v) {
				out <- v
			}
		}
	}()

	return out
}

func Send[T any](ctx context.Context, ch chan<- T, v T) error {
	select {
	case ch <- v:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func RecvErr(ctx context.Context, ch <-chan error) error {
	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func Recv[T any](ctx context.Context, ch <-chan T) (T, error) {
	select {
	case v := <-ch:
		return v, nil
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	}
}

// WaitAllDone returns a channel that is closed when all provided channels are closed
// or when ctx is done (whichever happens first).
func WaitAllDone(ctx context.Context, chans ...<-chan struct{}) <-chan struct{} {
	done := make(chan struct{})

	if len(chans) == 0 {
		close(done)
		return done
	}

	go func() {
		defer close(done)
		for _, ch := range chans {
			select {
			case <-ctx.Done():
				return
			case <-ch:
			}
		}
	}()

	return done
}
