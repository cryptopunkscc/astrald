package sig

import (
	"context"
	"sync"
)

type Task func(context.Context)

// Workers spawns count workers that to work on the queue. Returns a channel that will be closed when all
// workers return.
func Workers(ctx context.Context, queue <-chan Task, count int) <-chan struct{} {
	var wg sync.WaitGroup
	var sig = New()

	wg.Add(count)
	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case task := <-queue:
					task(ctx)
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(sig)
	}()

	return sig
}
