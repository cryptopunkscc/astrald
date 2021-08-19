package apphost

import (
	"context"
	"io"
	"sync"
)

func join(ctx context.Context, left, right io.ReadWriteCloser) error {
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

	go func() {
		<-ctx.Done()
		right.Close()
	}()

	wg.Wait()
	return nil
}
