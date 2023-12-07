package sig

import (
	"errors"
	"slices"
	"sync"
)

type Set[T comparable] struct {
	items []T
	mu    sync.RWMutex
}

func (set *Set[T]) Add(item T) error {
	set.mu.Lock()
	defer set.mu.Unlock()

	for _, i := range set.items {
		if i == item {
			return errors.New("already added")
		}
	}

	set.items = append(set.items, item)

	return nil
}

func (set *Set[T]) Remove(item T) error {
	set.mu.Lock()
	defer set.mu.Unlock()

	for idx, i := range set.items {
		if i == item {
			set.items = append(set.items[:idx], set.items[idx+1:]...)
			return nil
		}
	}

	set.items = append(set.items, item)

	return errors.New("not found")
}

func (set *Set[T]) Clone() []T {
	set.mu.RLock()
	defer set.mu.RUnlock()

	return slices.Clone(set.items)
}
