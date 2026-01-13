package sig

import (
	"context"
	"sync"
)

// ChanToArray drains ch until it is closed and returns all values in arrival order.
func ChanToArray[T any](ch <-chan T) (arr []T) {
	for i := range ch {
		arr = append(arr, i)
	}
	return
}

// ArrayToChan converts arr into a closed, buffered channel containing all elements.
func ArrayToChan[T any](arr []T) chan T {
	var ch = make(chan T, len(arr))
	for _, i := range arr {
		ch <- i
	}
	close(ch)
	return ch
}

// ChanFanIn merges multiple input channels into a single output channel.
//
// Semantics / invariants:
//   - Per-input ordering is preserved.
//   - No ordering guarantees are made across different inputs.
//   - The output channel is closed when all input channels are closed.
//   - If ctx is canceled, forwarding stops as soon as goroutines observe ctx.Done().
//     The output channel is then closed after all per-input goroutines exit.
//   - If streams is empty, returns a closed channel.
//
// Notes:
//   - If ctx is canceled, buffered values already sitting in input channels may be
//     dropped (the function prioritizes cancellation over draining).
//   - If any stream is nil, the goroutine for that stream will block forever unless
//     ctx is canceled.
func ChanFanIn[T any](
	ctx context.Context,
	streams ...<-chan T,
) <-chan T {
	out := make(chan T, len(streams)*2)

	if len(streams) == 0 {
		close(out)
		return out
	}

	var wg sync.WaitGroup
	wg.Add(len(streams))

	for _, s := range streams {
		stream := s
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case v, ok := <-stream:
					if !ok {
						return
					}
					select {
					case out <- v:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// ChanCollectAll merges streams with ChanFanIn and drains them into a slice.
func ChanCollectAll[T any](
	ctx context.Context,
	streams ...<-chan T,
) []T {
	return ChanToArray(ChanFanIn(ctx, streams...))
}

// MapChan transforms values from src into a new channel using fn.
// The output channel is closed when src is closed or ctx is canceled.
func MapChan[A any, B any](
	ctx context.Context,
	src <-chan A,
	fn func(A) B,
) <-chan B {
	out := make(chan B, 16)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-src:
				if !ok {
					return
				}
				select {
				case out <- fn(v):
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out
}
