package sig

import "sync"

// Value is a thread-safe generic value container
type Value[T comparable] struct {
	v  T
	mu sync.RWMutex
}

// Set sets the value to v
func (m *Value[V]) Set(v V) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.v = v
}

// Get returns the value
func (m *Value[V]) Get() V {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.v
}

// Swap sets the value to new only if the current value is old.
func (m *Value[T]) Swap(old T, new T) (final T, swapped bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.v != old {
		return m.v, false
	}

	m.v = new

	return m.v, true
}
