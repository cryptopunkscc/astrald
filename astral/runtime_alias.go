package astral

import (
	"encoding/json"
	"fmt"
	"io"
)

// RuntimeAlias is the live encoder/decoder for a BlueprintAlias. It wraps the underlying
// primitive Object and reports the alias name as its ObjectType — the wire counterpart of
// what RuntimeObject does for runtime-registered struct Blueprints. Originating peers that
// link the alias's Go type still get the typed Go value back via the compile-time prototype
// branch in New; this carrier exists for peers that learned the type only via blueprint sync.
type RuntimeAlias struct {
	a     *BlueprintAlias
	value Object
}

// NewRuntimeAlias returns a RuntimeAlias whose value is initialized to the zero value of
// the underlying primitive.
func NewRuntimeAlias(a *BlueprintAlias) (*RuntimeAlias, error) {
	if a == nil {
		return &RuntimeAlias{}, nil
	}
	err := validateAlias(a)
	if err != nil {
		return nil, err
	}

	v := New(a.Underlying.String())
	if v == nil {
		return nil, fmt.Errorf("%w: %s", ErrBlueprintNotFound, a.Underlying)
	}
	return &RuntimeAlias{a: a, value: v}, nil
}

// GetRuntimeAlias returns a fresh RuntimeAlias backed by this BlueprintAlias.
func (a *BlueprintAlias) GetRuntimeAlias() (*RuntimeAlias, error) { return NewRuntimeAlias(a) }

// astral:blueprint-ignore
func (ra *RuntimeAlias) ObjectType() string {
	if ra.a == nil {
		return ""
	}
	return ra.a.Type.String()
}

// Underlying returns the wrapped primitive Object. Callers may type-assert it to the
// specific astral primitive (e.g. *Uint8) named by the BlueprintAlias.
func (ra *RuntimeAlias) Underlying() Object { return ra.value }

func (ra *RuntimeAlias) WriteTo(w io.Writer) (int64, error) {
	if ra.a == nil {
		return 0, nil
	}
	return ra.value.WriteTo(w)
}

func (ra *RuntimeAlias) ReadFrom(r io.Reader) (int64, error) {
	if ra.a == nil {
		return 0, nil
	}
	return ra.value.ReadFrom(r)
}

// MarshalJSON serializes the alias as its underlying value — the alias name is metadata
// available via ObjectType and does not appear in the JSON shape. A nil binding marshals
// to null.
func (ra *RuntimeAlias) MarshalJSON() ([]byte, error) {
	if ra.a == nil {
		return jsonNull, nil
	}
	return json.Marshal(ra.value)
}

// UnmarshalJSON decodes the JSON payload into the underlying primitive. The receiver must
// already be bound to a BlueprintAlias (via NewRuntimeAlias or astral.New(typeName)) —
// without it there's no underlying type to instantiate.
func (ra *RuntimeAlias) UnmarshalJSON(data []byte) error {
	if ra.a == nil {
		return fmt.Errorf("RuntimeAlias.UnmarshalJSON: no BlueprintAlias bound")
	}
	return json.Unmarshal(data, &ra.value)
}
