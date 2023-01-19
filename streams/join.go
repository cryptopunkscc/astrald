package streams

import (
	"io"
	"sync"
)

// Join copies from each stream to the other until any of them closes and then returns nil.
func Join(left, right io.ReadWriteCloser) error {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		io.Copy(left, right)
		left.Close()
		wg.Done()
	}()

	go func() {
		io.Copy(right, left)
		right.Close()
		wg.Done()
	}()

	wg.Wait()
	return nil
}
