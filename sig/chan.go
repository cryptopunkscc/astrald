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
func ArrayToChan[T any](arr []T) chan T {
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
