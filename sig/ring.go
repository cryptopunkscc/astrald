package sig

import (
	"errors"
	"sync"
)

// Ring is a thread-safe ring buffer.
type Ring[T any] struct {
	items    []T
	head     int
	tail     int
	size     int
	capacity int

	mu sync.Mutex // protects all fields
}

// NewRing returns a new Ring with given capacity.
func NewRing[T any](capacity int) (*Ring[T], error) {
	if capacity <= 0 {
		return nil, errors.New("capacity must be positive")
	}
	return &Ring[T]{
		items:    make([]T, capacity),
		capacity: capacity,
	}, nil
}

// Push adds an item to the buffer. Returns true and the oldest value if the buffer was full.
func (ring *Ring[T]) Push(item T) (oldest T, overwritten bool) {
	ring.mu.Lock()
	defer ring.mu.Unlock()

	old := ring.items[ring.head]
	ring.items[ring.head] = item
	ring.head = (ring.head + 1) % ring.capacity

	if ring.size < ring.capacity {
		ring.size++
		var zero T
		return zero, false
	}

	ring.tail = (ring.tail + 1) % ring.capacity
	return old, true
}

// Pop removes and returns the oldest item. Returns zero value and false if empty.
func (ring *Ring[T]) Pop() (T, bool) {
	ring.mu.Lock()
	defer ring.mu.Unlock()

	if ring.size == 0 {
		var zero T
		return zero, false
	}

	item := ring.items[ring.tail]
	var zero T
	ring.items[ring.tail] = zero // zero pointers to allow for gc
	ring.tail = (ring.tail + 1) % ring.capacity
	ring.size--

	return item, true
}

// Peek returns the oldest item without removing it. Returns zero value and false if empty.
func (ring *Ring[T]) Peek() (T, bool) {
	ring.mu.Lock()
	defer ring.mu.Unlock()

	if ring.size == 0 {
		var zero T
		return zero, false
	}
	return ring.items[ring.tail], true
}

// Len returns the current number of elements.
func (ring *Ring[T]) Len() int {
	ring.mu.Lock()
	defer ring.mu.Unlock()
	return ring.size
}

// Cap returns the maximum capacity.
func (ring *Ring[T]) Cap() int {
	return ring.capacity
}

// Clear clears the buffer while keeping capacity.
func (ring *Ring[T]) Clear() {
	ring.mu.Lock()
	defer ring.mu.Unlock()

	ring.head = 0
	ring.tail = 0
	ring.size = 0
	for i := range ring.items {
		ring.items[i] = *new(T)
	}
}

// Clone returns a copy of current contents in order (oldest first).
func (ring *Ring[T]) Clone() (clone []T) {
	ring.mu.Lock()
	defer ring.mu.Unlock()

	clone = make([]T, 0, ring.size)
	pos := ring.tail
	for i := 0; i < ring.size; i++ {
		clone = append(clone, ring.items[pos])
		pos = (pos + 1) % ring.capacity
	}
	return clone
}
