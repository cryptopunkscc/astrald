package tree

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/sig"
)

// Value wraps an astral.Object type with type-safe access.
type Value[T astral.Object] struct {
	node     Node
	onChange func(T)
	value    *sig.Value[astral.Object]
	queue    *sig.Queue[T]
}

var _ astral.Object = &Value[astral.Object]{}

func (value *Value[T]) ObjectType() string {
	var zero T
	return zero.ObjectType()
}

func newValue[T astral.Object](node Node, onChange func(T)) *Value[T] {
	return &Value[T]{
		node:     node,
		onChange: onChange,
		value:    &sig.Value[astral.Object]{},
		queue:    &sig.Queue[T]{},
	}
}

func (value *Value[T]) Bind(ctx *astral.Context, node Node, configFunc ...BindConfigFunc[T]) error {
	b, err := Bind[T](ctx, node, configFunc...)
	if err != nil {
		*value = Value[T]{value: &sig.Value[astral.Object]{}}
		return err
	}
	*value = *b
	return err
}

// Get returns the current value as type T.
func (value *Value[T]) Get() (T, error) {
	var zero T
	v := value.value.Get()
	if v == nil {
		return zero, nil
	}
	if typed, ok := v.(T); ok {
		return typed, nil
	}
	return zero, ErrTypeMismatch
}

// Set updates the value.
func (value *Value[T]) Set(ctx *astral.Context, v T) error {
	return value.node.Set(ctx, v)
}

func (value Value[T]) WriteTo(writer io.Writer) (n int64, err error) {
	v := value.value.Get()
	if v == nil {
		return 0, errors.New("nil value")
	}
	return v.WriteTo(writer)
}

func (value *Value[T]) ReadFrom(reader io.Reader) (n int64, err error) {
	var obj T
	n, err = obj.ReadFrom(reader)
	if err != nil {
		return
	}

	value.value.Set(obj)
	return
}

func (value Value[T]) MarshalJSON() ([]byte, error) {
	obj := value.value.Get()

	if m, ok := obj.(json.Marshaler); ok {
		return m.MarshalJSON()
	}

	return nil, errors.New("object does not implement json.Marshaler")
}

func (value *Value[T]) UnmarshalJSON(bytes []byte) error {
	var obj T
	err := json.Unmarshal(bytes, &obj)
	if err != nil {
		return err
	}

	value.value.Set(obj)

	return nil
}

func (value *Value[T]) Subscribe(ctx *astral.Context) <-chan T {
	return sig.Subscribe(ctx, value.queue)
}
