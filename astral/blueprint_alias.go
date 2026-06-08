package astral

import (
	"fmt"
	"io"
)

var _ Object = (*BlueprintAlias)(nil)

// Aliasable lets a compile-time Object prototype declare itself as a primitive alias so
// the registry can derive a BlueprintAlias for it on the fly — the Go-side counterpart of
// `type Foo astral.Uint8`. Apps add a single method:
//
//	func (*Mode) UnderlyingPrimitive() string { return "uint8" }
//
// `astral.Add(&Mode{})` is then enough: local New() returns the typed Go value, and
// AllAliases derives a BlueprintAlias for sync without a separate RegisterAlias call.
type Aliasable interface {
	UnderlyingPrimitive() string
}

// BlueprintAlias is the runtime descriptor for a named primitive alias — the schema
// counterpart of `type Foo astral.Uint8` in Go. Wire bytes for an alias-typed value
// are identical to the bytes of its Underlying primitive; the alias name lives in the
// registry, not in the payload.
//
// Aliases over struct types are intentionally out of scope: Blueprint already covers
// that shape via Fields, and the use case (newtype-over-struct) has no wire benefit.
type BlueprintAlias struct {
	Type       String16 // e.g. "mod.nearby.mode"
	Underlying String16 // primitive name from primitiveAllowlist (e.g. "uint8")
}

func (*BlueprintAlias) ObjectType() string { return "astral.blueprint.alias" }

func (a BlueprintAlias) WriteTo(w io.Writer) (int64, error)   { return Objectify(&a).WriteTo(w) }
func (a *BlueprintAlias) ReadFrom(r io.Reader) (int64, error) { return Objectify(a).ReadFrom(r) }

// NewBlueprintAlias returns a BlueprintAlias with the given type and underlying primitive.
func NewBlueprintAlias(typeName, underlying string) *BlueprintAlias {
	return &BlueprintAlias{Type: String16(typeName), Underlying: String16(underlying)}
}

// AliasOf derives a BlueprintAlias from an Object that implements Aliasable — the
// counterpart to BlueprintOf for struct prototypes. Returns ErrBlueprintInvalid if the
// Object doesn't satisfy Aliasable or its declared Underlying isn't on the primitive
// allowlist.
func AliasOf(o Object) (*BlueprintAlias, error) {
	a, ok := o.(Aliasable)
	if !ok {
		return nil, fmt.Errorf("%w: AliasOf: %T does not implement Aliasable",
			ErrBlueprintInvalid, o)
	}
	bp := &BlueprintAlias{
		Type:       String16(o.ObjectType()),
		Underlying: String16(a.UnderlyingPrimitive()),
	}
	if err := validateAlias(bp); err != nil {
		return nil, fmt.Errorf("AliasOf: %w", err)
	}
	return bp, nil
}

// validateAlias enforces the same name discipline as Blueprint (ASCII, ≤MaxBlueprintNameLen,
// non-empty Type) and pins Underlying to the primitive allowlist.
func validateAlias(a *BlueprintAlias) error {
	if a == nil || a.Type.String() == "" {
		return fmt.Errorf("%w: empty Type", ErrBlueprintInvalid)
	}
	if !isASCII(a.Type.String()) {
		return fmt.Errorf("%w: Type must be ASCII", ErrBlueprintInvalid)
	}
	if len(a.Type) > MaxBlueprintNameLen {
		return fmt.Errorf("%w: Type exceeds %d bytes", ErrBlueprintInvalid, MaxBlueprintNameLen)
	}
	if !isAllowedPrimitive(a.Underlying.String()) {
		return fmt.Errorf("%w: Underlying %q not in primitive allowlist", ErrBlueprintInvalid, a.Underlying)
	}
	return nil
}

func init() { _ = Add(&BlueprintAlias{}) }
