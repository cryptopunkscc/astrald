package sig

import (
	"errors"
	"slices"
	"sync"
)

type Set[T comparable] struct {
	items []T
	dup   map[T]struct{}
	mu    sync.RWMutex
}

type ErrDuplicateItem struct {
	Index int
}

func (e ErrDuplicateItem) Error() string { return "duplicate item" }

func (set *Set[T]) Add(items ...T) error {
	set.mu.Lock()
	defer set.mu.Unlock()

	if set.dup == nil {
		set.dup = make(map[T]struct{})
	}

	var errs []error

	for idx, item := range items {
		if _, found := set.dup[item]; found {
			errs = append(errs, ErrDuplicateItem{Index: idx})
			continue
		}
		set.items = append(set.items, item)
		set.dup[item] = struct{}{}
	}

	return errors.Join(errs...)
}

func (set *Set[T]) Remove(item T) error {
	set.mu.Lock()
	defer set.mu.Unlock()

	if set.dup == nil {
		return errors.New("not found")
	}

	if _, found := set.dup[item]; !found {
		return errors.New("not found")
	}

	delete(set.dup, item)

	for idx, i := range set.items {
		if i == item {
			set.items = append(set.items[:idx], set.items[idx+1:]...)
			break
		}
	}

	return nil
}

func (set *Set[T]) Clear() error {
	set.mu.Lock()
	defer set.mu.Unlock()

	set.items = nil
	set.dup = make(map[T]struct{})

	return nil
}

func (set *Set[T]) Contains(item T) bool {
	set.mu.RLock()
	defer set.mu.RUnlock()

	_, found := set.dup[item]
	return found
}

func (set *Set[T]) Clone() []T {
	set.mu.RLock()
	defer set.mu.RUnlock()

	return slices.Clone(set.items)
}

func (set *Set[T]) Count() int {
	set.mu.RLock()
	defer set.mu.RUnlock()

	return len(set.items)
}

func (set *Set[T]) Sort(cmp func(a, b T) int) {
	set.mu.Lock()
	defer set.mu.Unlock()

	slices.SortFunc(set.items, cmp)
}

func (set *Set[T]) Select(fn func(a T) bool) (s []T) {
	set.mu.RLock()
	defer set.mu.RUnlock()

	for _, item := range set.items {
		if fn(item) {
			s = append(s, item)
		}
	}

	return
}
