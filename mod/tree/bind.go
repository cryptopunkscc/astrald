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
	onChange func(T)
	value    sig.Value[astral.Object]
}

// Bind creates a binding to a tree node.
func Bind[T astral.Object](ctx *astral.Context, node Node, configFunc ...BindConfigFunc[T]) (*Binding[T], error) {
	var err error
	var config = makeBindConfig(configFunc...)

	// query the node if an additional path was configured
	if config.Path != "" {
		node, err = Query(ctx, node, config.Path, true)
		if err != nil {
			return nil, err
		}
	}

	// fork our internal context
	ctx, cancel := ctx.WithCancel()

	// try to follow the node and set the default value if necessary
	ch, err := Follow[T](ctx, node)
	switch {
	case err == nil:
	case errors.Is(err, &ErrNodeHasNoValue{}):
		// set the default value
		err = node.Set(ctx, config.DefaultValue)
		if err != nil {
			cancel()
			return nil, err
		}

		// try to follow one more time
		ch, err = Follow[T](ctx, node)
		if err != nil {
			cancel()
			return nil, err
		}

	default:
		cancel()
		return nil, err
	}

	// create the binding
	binding := &Binding[T]{
		node:     node,
		onChange: config.OnChange,
	}

	// wait for the initial value
	select {
	case v := <-ch:
		binding.value.Set(v)
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// subscribe to changes
	go func() {
		defer cancel()
		for obj := range ch {
			binding.value.Set(obj)

			if binding.onChange != nil {
				binding.onChange(obj)
			}
		}
	}()

	return binding, nil
}

func BindPath[T astral.Object](ctx *astral.Context, node Node, path string, configFunc ...BindConfigFunc[T]) (*Binding[T], error) {
	node, err := Query(ctx, node, path, true)
	if err != nil {
		return nil, err
	}

	return Bind(ctx, node, configFunc...)
}

// Get returns the current value as type T.
func (binding *Binding[T]) Get() (T, error) {
	var zero T
	v := binding.value.Get()
	if v == nil {
		return zero, nil
	}
	if typed, ok := v.(T); ok {
		return typed, nil
	}
	return zero, ErrTypeMismatch
}

// Set updates the value.
func (binding *Binding[T]) Set(ctx *astral.Context, v T) error {
	return binding.node.Set(ctx, v)
}
