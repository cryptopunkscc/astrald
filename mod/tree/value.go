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
	node    Node
	cached  *sig.Value[astral.Object]
	queue   *sig.Queue[T]
	NoInit  bool
	NoLocal bool
}

var _ astral.Object = &Value[astral.Object]{}

func (value *Value[T]) ObjectType() string {
	var zero T
	return zero.ObjectType()
}

// Bind binds the value to the node until the context is canceled.
func (value *Value[T]) Bind(ctx *astral.Context, node Node) error {
	// follow the node
	updates, errPtr := Follow[T](ctx, node)
	switch {
	case *errPtr == nil:
	case errors.Is(*errPtr, &ErrNoValue{}):
		if value.NoInit {
			return *errPtr
		}

		// set zero value
		err := value.setZero(ctx)
		if err != nil {
			return err
		}

		// follow again
		updates, errPtr = Follow[T](ctx, node)
	}
	if *errPtr != nil {
		return *errPtr
	}

	if value.cached == nil {
		value.cached = &sig.Value[astral.Object]{}
	}
	if value.queue == nil {
		value.queue = &sig.Queue[T]{}
	}

	// set the initial value
	value.update(<-updates)

	// subscribe to changes
	go func() {
		for obj := range updates {
			value.update(obj)
		}
		// TODO: try to reconnect on recoverable errors?
	}()

	return nil
}

// BindPath is a convenience function that queries the node and calls Bind.
func (value *Value[T]) BindPath(ctx *astral.Context, node Node, path string, create bool) (err error) {
	node, err = Query(ctx, node, path, create)
	if err != nil {
		return err
	}

	return value.Bind(ctx, node)
}

// Get returns the currently held object
func (value *Value[T]) Get() (val T, err error) {
	if value.cached == nil {
		return val, nil
	}

	// get cached value
	v := value.cached.Get()
	if v == nil {
		return val, nil
	}

	// cast the value
	if typed, ok := v.(T); ok {
		return typed, nil
	}

	return val, ErrTypeMismatch
}

// Set updates the value.
func (value *Value[T]) Set(ctx *astral.Context, v T) error {
	// set the value remotely
	err := value.node.Set(ctx, v)
	if err == nil {
		return nil
	}

	// if no local update is allowed, return the error
	if value.NoLocal {
		return err
	}

	if value.cached == nil {
		value.cached = &sig.Value[astral.Object]{}
	}
	if value.queue == nil {
		value.queue = &sig.Queue[T]{}
	}

	// update the value locally
	value.update(v)

	return err
}

// Follow returns a channel that emits the current value and all updates.
func (value *Value[T]) Follow(ctx *astral.Context) <-chan T {
	var out = make(chan T, 1)
	val, err := value.Get()
	if err != nil {
		close(out)
		return out
	}

	// send the initial value
	select {
	case <-ctx.Done():
		close(out)
		return out
	case out <- val:
	}

	// subscribe to updates
	go func() {
		defer close(out)
		for val := range sig.Subscribe(ctx, value.queue) {
			select {
			case <-ctx.Done():
				return
			case out <- val:
			}
		}
	}()

	return out
}

// update updates the cached value and pushes it to the queue
func (value *Value[T]) update(val T) {
	value.cached.Set(val)
	value.queue = value.queue.Push(val)
}

func (value *Value[T]) setZero(ctx *astral.Context) error {
	var zero T
	return value.node.Set(ctx, zero)
}

func (value Value[T]) WriteTo(writer io.Writer) (n int64, err error) {
	v := value.cached.Get()
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

	value.cached.Set(obj)
	return
}

func (value Value[T]) MarshalJSON() ([]byte, error) {
	if value.cached == nil {
		return json.Marshal(nil)
	}

	obj := value.cached.Get()

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

	value.cached.Set(obj)

	return nil
}
