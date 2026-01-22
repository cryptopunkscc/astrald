package tree

import (
	"errors"
	"reflect"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/sig"
)

var ErrTypeMismatch = errors.New("binding type mismatch")

// Binding wraps a BindingIface with type-safe access.
type Binding[T astral.Object] struct {
	mod      Module
	node     Node
	cancel   func()
	onChange func(astral.Object)
	value    sig.Value[astral.Object]
}

// Bind creates a binding to a path that tracks value changes.
// If defaultValue is non-nil and no value exists, it sets the default.
// onChange can be nil if no callback is needed.
func Bind[T astral.Object](mod Module, path string, configFunc ...BindConfigFunc[T]) (*Binding[T], error) {
	ctx := mod.Context()
	config := MakeBindConfig(configFunc...)

	node, err := Query(ctx, mod.Root(), path, true)
	if err != nil {
		return nil, err
	}

	subCtx, cancel := ctx.WithCancel()
	ch, err := node.Get(subCtx, true)
	switch {
	case err == nil:
	case errors.Is(err, &ErrNodeHasNoValue{}):
		// set the default value
		err = node.Set(ctx, config.DefaultValue)
		if err != nil {
			return nil, err
		}

		// try to follow one more time
		ch, err = node.Get(subCtx, true)
		if err != nil {
			return nil, err
		}

	default:
		cancel()
		return nil, err
	}

	b := &Binding[T]{
		mod:      mod,
		node:     node,
		cancel:   cancel,
		onChange: config.OnChange,
	}

	go func() {
		for obj := range ch {
			b.value.Set(obj)

			if b.onChange != nil {
				b.onChange(obj)
			}
		}
	}()

	mod.RegisterBinding(path, b)

	return b, nil
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

type BindConfigFunc[T astral.Object] func(*BindConfig[T])

type BindConfig[T astral.Object] struct {
	DefaultValue T
	OnChange     func(astral.Object)
}

func WithDefaultValue[T astral.Object](v T) BindConfigFunc[T] {
	return func(cfg *BindConfig[T]) { cfg.DefaultValue = v }
}

func WithOnChange[T astral.Object](f func(T)) BindConfigFunc[T] {
	return func(cfg *BindConfig[T]) {
		cfg.OnChange = func(object astral.Object) {
			typed, ok := object.(T)
			if ok {
				f(typed)
			}
		}
	}
}

func MakeBindConfig[T astral.Object](fns ...BindConfigFunc[T]) BindConfig[T] {
	var cfg = BindConfig[T]{}
	var t T

	var v = reflect.ValueOf(t)
	if v.Kind() == reflect.Ptr {
		t = reflect.New(v.Type().Elem()).Interface().(T)
	}

	if o := astral.New(t.ObjectType()); o != nil {
		cfg.DefaultValue = o.(T)
	}

	for _, fn := range fns {
		fn(&cfg)
	}
	return cfg
}
