package tree

import (
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
)

// Value is a thread-safe wrapper that holds an up-to-date value from the tree.
// It automatically follows changes to a tree path and keeps the local value synchronized.
type Value[T astral.Object] struct {
	mu    sync.RWMutex
	value T
}

// NewValue creates a new Value wrapper with an initial value.
func NewValue[T astral.Object](initial T) *Value[T] {
	return &Value[T]{value: initial}
}

// Get returns the current value.
func (v *Value[T]) Get() T {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.value
}

// Set updates the value.
func (v *Value[T]) Set(value T) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.value = value
}

// Follow starts watching a tree node and updates the value when the node changes.
// It returns a stop function that should be called to stop following.
func (v *Value[T]) Follow(ctx *astral.Context, node Node) error {
	ch, err := Follow[T](ctx, node)
	if err != nil {
		return err
	}

	go func() {
		for val := range ch {
			v.Set(val)
		}
	}()

	return nil
}
