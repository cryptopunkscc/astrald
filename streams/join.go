package streams

import (
	"io"
	"sync"
)

// Join copies from each stream to the other until one of them reaches EOF or an error. Returns bytes written
// both ways, and a nil error if EOF was reached, or the first non-EOF error.
func Join(left, right io.ReadWriteCloser) (writtenL int64, writtenR int64, err error) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		defer left.Close()

		var errL error
		writtenL, errL = io.Copy(left, right)
		if err == nil {
			err = errL
		}
	}()

	go func() {
		defer wg.Done()
		defer right.Close()

		var errR error
		writtenR, errR = io.Copy(right, left)
		if err == nil {
			err = errR
		}
	}()

	wg.Wait()
	return
}
