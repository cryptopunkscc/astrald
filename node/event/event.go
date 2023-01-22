package event

import (
	"context"
	"github.com/cryptopunkscc/astrald/sig"
	"sync"
)

type Event interface{}

type Queue struct {
	mu     sync.Mutex
	queue  *sig.Queue
	parent *Queue
}

func (q *Queue) Emit(event Event) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.queue == nil {
		q.queue = &sig.Queue{}
	}

	q.queue = q.queue.Push(event)
	if q.parent != nil {
		q.parent.Emit(event)
	}
}

func (q *Queue) Subscribe(ctx context.Context) <-chan Event {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.queue == nil {
		q.queue = &sig.Queue{}
	}

	var ch = make(chan Event)

	go func() {
		defer close(ch)
		for e := range q.queue.Subscribe(ctx) {
			ch <- e
		}
	}()

	return ch
}

func (q *Queue) SetParent(parent *Queue) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q == parent {
		return
	}
	q.parent = parent
}

func (q *Queue) Parent() *Queue {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.parent
}
