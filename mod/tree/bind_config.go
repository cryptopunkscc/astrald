package tree

import (
	"reflect"

	"github.com/cryptopunkscc/astrald/astral"
)

// BindConfigFunc is a function that sets binding options.
type BindConfigFunc[T astral.Object] func(*bindConfig[T])

// DefaultValue sets the default value for the binding.
func DefaultValue[T astral.Object](v T) BindConfigFunc[T] {
	return func(cfg *bindConfig[T]) { cfg.DefaultValue = v }
}

// OnChange sets a callback called when the value changes.
func OnChange[T astral.Object](f func(T)) BindConfigFunc[T] {
	return func(cfg *bindConfig[T]) {
		cfg.OnChange = f
	}
}

// Path sets the path to query from the node.
func Path[T astral.Object](path string) BindConfigFunc[T] {
	return func(cfg *bindConfig[T]) { cfg.Path = path }
}

type bindConfig[T astral.Object] struct {
	DefaultValue T
	OnChange     func(T)
	Path         string
}

func makeBindConfig[T astral.Object](fns ...BindConfigFunc[T]) bindConfig[T] {
	var cfg = bindConfig[T]{}
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
