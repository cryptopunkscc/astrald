package sig

func ChanToArray[T any](ch chan T) (arr []T) {
	for i := range ch {
		arr = append(arr, i)
	}
	return
}
