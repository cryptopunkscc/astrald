package tree

import (
	"errors"

	"github.com/cryptopunkscc/astrald/astral"
)

var ErrTypeMismatch = errors.New("binding type mismatch")

// Binding represents a live connection to a tree path value.
type Binding interface {
	// Value returns the current value.
	Value() astral.Object

	// Set updates the value.
	Set(ctx *astral.Context, v astral.Object) error

	// Close stops tracking changes.
	Close()
}

// TypedBinding wraps a Binding with type-safe access.
type TypedBinding[T astral.Object] struct {
	Binding
}

// Typed wraps a Binding for type-safe access.
func Typed[T astral.Object](b Binding, err error) (*TypedBinding[T], error) {
	if err != nil {
		return nil, err
	}
	return &TypedBinding[T]{Binding: b}, nil
}

// Value returns the current value as type T.
func (tb *TypedBinding[T]) Value() (T, error) {
	var zero T
	v := tb.Binding.Value()
	if v == nil {
		return zero, nil
	}
	if typed, ok := v.(T); ok {
		return typed, nil
	}
	return zero, ErrTypeMismatch
}

// Set updates the value.
func (tb *TypedBinding[T]) Set(ctx *astral.Context, v T) error {
	return tb.Binding.Set(ctx, v)
}
