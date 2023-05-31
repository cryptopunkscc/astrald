package discovery

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/node/services"
	"sync"
	"sync/atomic"
)

type QueryWorkers struct {
	handler services.QueryHandlerFunc
	queue   chan *services.Query
	done    chan struct{}
	closed  atomic.Bool
}

func RunQueryWorkers(ctx context.Context, handler services.QueryHandlerFunc, count int) *QueryWorkers {
	w := &QueryWorkers{
		handler: handler,
		queue:   make(chan *services.Query),
		done:    make(chan struct{}),
	}

	var wg sync.WaitGroup
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case item := <-w.queue:
					if err := handler(ctx, item); err != nil {
						return
					}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		w.Close()
	}()

	return w
}

func (w *QueryWorkers) Enqueue(ctx context.Context, query *services.Query) error {
	if w.closed.Load() {
		return errors.New("closed")
	}
	select {
	case w.queue <- query:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (w *QueryWorkers) Close() error {
	if w.closed.CompareAndSwap(false, true) {
		close(w.queue)
		close(w.done)
	}
	return nil
}

func (w *QueryWorkers) Done() chan struct{} {
	return w.done
}
