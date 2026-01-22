package tree

import (
	"errors"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/sig"
)

var ErrTypeMismatch = errors.New("binding type mismatch")

// Binding wraps an astral.Object type with type-safe access.
type Binding[T astral.Object] struct {
	node     Node
	cancel   func()
	onChange func(T)
	value    sig.Value[astral.Object]
}

// Bind creates a binding to a tree node.
func Bind[T astral.Object](ctx *astral.Context, node Node, configFunc ...BindConfigFunc[T]) (*Binding[T], error) {
	var err error
	var config = makeBindConfig(configFunc...)

	if config.Path != "" {
		node, err = Query(ctx, node, config.Path, true)
		if err != nil {
			return nil, err
		}
	}

	subCtx, cancel := ctx.WithCancel()

	ch, err := Follow[T](subCtx, node)
	switch {
	case err == nil:
	case errors.Is(err, &ErrNodeHasNoValue{}):
		// set the default value
		err = node.Set(ctx, config.DefaultValue)
		if err != nil {
			return nil, err
		}

		// try to follow one more time
		ch, err = Follow[T](subCtx, node)
		if err != nil {
			return nil, err
		}

	default:
		cancel()
		return nil, err
	}

	b := &Binding[T]{
		node:     node,
		cancel:   cancel,
		onChange: config.OnChange,
	}

	select {
	case v := <-ch:
		b.value.Set(v)
	case <-subCtx.Done():
	}

	go func() {
		for obj := range ch {
			b.value.Set(obj)

			if b.onChange != nil {
				b.onChange(obj)
			}
		}
	}()

	return b, nil
}

func BindPath[T astral.Object](ctx *astral.Context, node Node, path string, configFunc ...BindConfigFunc[T]) (*Binding[T], error) {
	node, err := Query(ctx, node, path, true)
	if err != nil {
		return nil, err
	}

	return Bind(ctx, node, configFunc...)
}

// Get returns the current value as type T.
func (tb *Binding[T]) Get() (T, error) {
	var zero T
	v := tb.value.Get()
	if v == nil {
		return zero, nil
	}
	if typed, ok := v.(T); ok {
		return typed, nil
	}
	return zero, ErrTypeMismatch
}

// Set updates the value.
func (tb *Binding[T]) Set(ctx *astral.Context, v T) error {
	return tb.node.Set(ctx, v)
}

func (tb *Binding[T]) Close() error {
	tb.cancel()
	return nil
}
