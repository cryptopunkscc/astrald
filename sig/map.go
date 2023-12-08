package sig

import "sync"

// Map is a basic thread-safe map
type Map[K comparable, V any] struct {
	m  map[K]V
	mu sync.RWMutex
}

// Set sets the value of the key only if no other value is aready set. Returns true on success.
func (m *Map[K, V]) Set(key K, value V) (ok bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.m == nil {
		m.m = map[K]V{}
	}

	if _, found := m.m[key]; found {
		return false
	}

	m.m[key] = value

	return true
}

// Replace sets the value of the key. If key already had a value it will be returned and ok will
// be true. If no value was assigned, a zero value and false will be returned.
func (m *Map[K, V]) Replace(key K, value V) (old V, ok bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.m == nil {
		m.m = map[K]V{}
	}

	old, ok = m.m[key]

	m.m[key] = value

	return
}

// Get returns the value assigned to the key
func (m *Map[K, V]) Get(key K) (v V, ok bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.m == nil {
		return
	}

	v, ok = m.m[key]

	return v, ok
}

// Delete clears the value assigned to the key. Returns false if no value was assigned,
// deleted value and true otherwise.
func (m *Map[K, V]) Delete(key K) (old V, ok bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.m == nil {
		return
	}

	old, ok = m.m[key]
	if ok {
		delete(m.m, key)
	}

	return
}

// Len returns the number of keys in the map
func (m *Map[K, V]) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.m)
}

// Clone returns a clone of the underlying map
func (m *Map[K, V]) Clone() (c map[K]V) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	c = map[K]V{}

	for k, v := range m.m {
		c[k] = v
	}

	return
}

// Keys returns a list of the keys in the map
func (m *Map[K, V]) Keys() (keys []K) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for k, _ := range m.m {
		keys = append(keys, k)
	}

	return
}
