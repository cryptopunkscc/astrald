package sig

import (
	"context"
	"sync"
)

type Queue struct {
	wait chan struct{}
	next *Queue
	data interface{}
	mu   sync.Mutex
}

// Push adds a new value to the queue and notifies Waiters
func (q *Queue) Push(data interface{}) *Queue {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.next != nil {
		return q.next.Push(data)
	}

	if q.wait == nil {
		q.wait = make(chan struct{})
	}
	defer close(q.wait)

	q.data = data
	q.next = &Queue{}

	return q.next
}

// Next should only be called after Wait channel closes. It returns the next Queue element or nil on EOF.
func (q *Queue) Next() *Queue {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.next
}

// Close marks this element as EOF and notifies Waiters
func (q *Queue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.next != nil {
		q.next.Close()
		return
	}

	if q.wait == nil {
		q.wait = make(chan struct{})
	}
	defer close(q.wait)

	q.next = nil
}

// Wait returns a channel that will be closed when the value of this element is ready
func (q *Queue) Wait() <-chan struct{} {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.wait == nil {
		q.wait = make(chan struct{})
	}

	return q.wait
}

// Data returns nil of the value is not ready (if called before Wait closes), the value otherwise
func (q *Queue) Data() interface{} {
	return q.data
}

func (q *Queue) Follow(ctx context.Context) <-chan interface{} {
	ch := make(chan interface{})

	go func() {
		f := q
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				return
			case <-f.Wait():
				if f.Next() == nil {
					return
				}

				select {
				case ch <- f.data:
					f = f.Next()
				case <-ctx.Done():
					return
				}

			}
		}
	}()

	return ch
}
