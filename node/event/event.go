package event

import (
	"context"
	"github.com/cryptopunkscc/astrald/sig"
	"sync"
)

type Event interface{}

type Queue struct {
	mu     sync.Mutex
	queue  *sig.Queue[Event]
	parent *Queue
}

// Emit pushes an event onto the event queue.
func (q *Queue) Emit(event Event) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.queue == nil {
		q.queue = &sig.Queue[Event]{}
	}

	q.queue = q.queue.Push(event)
	if q.parent != nil {
		q.parent.Emit(event)
	}
}

// Subscribe returns a channel, which will receive events from the queue until the context ends. Channel will be
// closed afterwards.
func (q *Queue) Subscribe(ctx context.Context) <-chan Event {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.queue == nil {
		q.queue = &sig.Queue[Event]{}
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

// Handle will subscribe to the Queue for the duration of the context and will invoke fn for every
// element that matches fn's argument type. If fn returns an error, Handle stops and retruns the error.
func Handle[EventType Event](ctx context.Context, q *Queue, fn func(EventType) error) error {
	return sig.Handle(ctx, q.queue, fn)
}

// SetParent sets the parent queue. All events emitted by this queue are propagated to the parent.
func (q *Queue) SetParent(parent *Queue) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q == parent {
		return
	}
	q.parent = parent
}

// Parent returns the parent queue. All events emitted by this queue are propagated to the parent.
func (q *Queue) Parent() *Queue {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.parent
}
