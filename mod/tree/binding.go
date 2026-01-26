package tree

import (
	"encoding/json"
	"errors"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/sig"
)

// Value wraps an astral.Object type with type-safe access.
type Value[T astral.Object] struct {
	node     Node
	onChange func(T)
	value    *sig.Value[astral.Object]
}

func (binding *Value[T]) Bind(ctx *astral.Context, node Node, configFunc ...BindConfigFunc[T]) error {
	b, err := Bind[T](ctx, node, configFunc...)
	if err != nil {
		*binding = Value[T]{value: &sig.Value[astral.Object]{}}
		return err
	}
	*binding = *b
	return err
}

func (binding Value[T]) MarshalJSON() ([]byte, error) {
	obj := binding.value.Get()

	if m, ok := obj.(json.Marshaler); ok {
		return m.MarshalJSON()
	}

	return nil, errors.New("object does not implement json.Marshaler")
}

func (binding *Value[T]) UnmarshalJSON(bytes []byte) error {
	var obj T
	err := json.Unmarshal(bytes, &obj)
	if err != nil {
		return err
	}

	binding.value.Set(obj)

	return nil
}
