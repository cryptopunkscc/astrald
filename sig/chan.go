package sig

func ChanToArray[T any](ch <-chan T) (arr []T) {
	for i := range ch {
		arr = append(arr, i)
	}
	return
}

func ArrayToChan[T any](arr []T) chan T {
	var ch = make(chan T, len(arr))
	for _, i := range arr {
		ch <- i
	}
	close(ch)
	return ch
}
