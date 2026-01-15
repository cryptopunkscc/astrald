package sig

import "sync"

// DedupQueue is a FIFO queue with integrated deduplication.
// Items are enqueued only if not already present in the queue.
// Dequeue blocks until an item is available or the queue is closed.
// All methods are safe for concurrent use.
type DedupQueue[T comparable] struct {
	mu      sync.Mutex
	cond    *sync.Cond
	items   []T
	inQueue map[T]struct{} // O(1) membership check
	closed  bool
}

func NewDedupQueue[T comparable]() *DedupQueue[T] {
	q := &DedupQueue[T]{
		items:   make([]T, 0, 32),
		inQueue: make(map[T]struct{}),
	}
	q.cond = sync.NewCond(&q.mu)
	return q
}

// Enqueue adds an item to the queue if not already present.
// Returns true if the item was added (not a duplicate).
// Returns false if the item is already queued or the queue is closed.
func (q *DedupQueue[T]) Enqueue(item T) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return false
	}

	if _, exists := q.inQueue[item]; exists {
		return false // duplicate
	}

	q.items = append(q.items, item)
	q.inQueue[item] = struct{}{}
	q.cond.Signal()

	return true
}

// Dequeue removes and returns the next item from the queue.
// Blocks if the queue is empty until an item is available or the queue is closed.
// Returns (zero value, false) if the queue is closed.
func (q *DedupQueue[T]) Dequeue() (T, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for {
		if q.closed {
			var zero T
			return zero, false
		}

		if len(q.items) > 0 {
			item := q.items[0]
			q.items = q.items[1:]
			delete(q.inQueue, item)
			return item, true
		}

		q.cond.Wait()
	}
}

func (q *DedupQueue[T]) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}

// Close marks the queue as closed and wakes all blocked dequeuers.
func (q *DedupQueue[T]) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return
	}

	q.closed = true
	q.cond.Broadcast() // wake all waiting dequeuers
}
