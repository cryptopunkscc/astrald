package tree

import (
	"errors"

	"github.com/cryptopunkscc/astrald/astral"
)

var ErrTypeMismatch = errors.New("binding type mismatch")

// Bind creates a binding to a tree node.
func Bind[T astral.Object](ctx *astral.Context, node Node, configFunc ...BindConfigFunc[T]) (*Value[T], error) {
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
	binding := newValue(node, config.OnChange)

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
			binding.queue = binding.queue.Push(obj)

			if binding.onChange != nil {
				binding.onChange(obj)
			}
		}
	}()

	return binding, nil
}

func BindPath[T astral.Object](ctx *astral.Context, node Node, path string, configFunc ...BindConfigFunc[T]) (*Value[T], error) {
	node, err := Query(ctx, node, path, true)
	if err != nil {
		return nil, err
	}

	return Bind(ctx, node, configFunc...)
}
