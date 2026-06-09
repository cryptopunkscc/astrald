package astral

import (
	"fmt"
	"io"
	"strings"
)

var _ Object = (*Blueprint)(nil)

// Blueprint describes a runtime-registered Object type by name. A Blueprint is one of two
// kinds, mutually exclusive:
//
//   - struct kind: Fields lists the ordered slots; Underlying is "".
//   - alias kind:  Underlying names a primitive from primitiveAllowlist (e.g. "uint8");
//     Fields is empty. Wire bytes for an alias-typed value are identical to the bytes of
//     its Underlying primitive — the alias name lives only in the registry.
//
// Aliases over struct types are intentionally out of scope: struct-kind Blueprint already
// covers that shape via Fields, and the use case (newtype-over-struct) has no wire benefit.
type Blueprint struct {
	Type       String16
	Fields     []Field
	Underlying String16
}

func (*Blueprint) ObjectType() string { return "astral.blueprint" }

// BlueprintKind classifies a Blueprint as struct-shaped (Fields) or alias-shaped (Underlying).
type BlueprintKind uint8

const (
	BlueprintKindStruct BlueprintKind = iota
	BlueprintKindAlias
)

// Kind reports whether the Blueprint describes a struct or a primitive alias.
// A non-empty Underlying selects alias kind; empty Underlying is struct kind (Fields may
// be empty — that's the empty-struct case, distinct from an alias).
func (bp *Blueprint) Kind() BlueprintKind {
	if bp.Underlying.String() != "" {
		return BlueprintKindAlias
	}
	return BlueprintKindStruct
}

// NewBlueprint returns a struct-kind Blueprint with the given type name and fields.
func NewBlueprint(typeName string, fields ...Field) *Blueprint {
	return &Blueprint{Type: String16(typeName), Fields: append([]Field{}, fields...)}
}

// NewBlueprintAlias returns an alias-kind Blueprint with the given type name and underlying
// primitive. Underlying must be on the primitive allowlist; validation defers to register time.
func NewBlueprintAlias(typeName, underlying string) *Blueprint {
	return &Blueprint{Type: String16(typeName), Underlying: String16(underlying)}
}

func (bp Blueprint) WriteTo(w io.Writer) (int64, error)   { return Objectify(&bp).WriteTo(w) }
func (bp *Blueprint) ReadFrom(r io.Reader) (int64, error) { return Objectify(bp).ReadFrom(r) }

var _ Object = (*Field)(nil)

// Spec describes the shape of one Field on the wire. Implementers are the closed set of Spec
// carriers: *PrimitiveSpec, *RefSpec, *SliceSpec, *ArraySpec, *MapSpec, *PtrSpec, *ObjectSpec.
// ReferencedType returns the Blueprint name this Spec depends on for closure validation, or ""
// when self-contained (primitive, heterogeneous container, ObjectSpec).
type Spec interface {
	Object
	ReferencedType() string
}

// Field is a single named slot inside a Blueprint. Spec is one of the Spec carriers
// (*PrimitiveSpec, *RefSpec, *SliceSpec, *ArraySpec, *MapSpec, *PtrSpec, *ObjectSpec) and is
// encoded polymorphically (type tag + payload) by Encode/Decode. Spec wire bytes are tag-less
// on their own; the discriminator comes from this interface-typed slot.
type Field struct {
	Name String16
	Spec Spec
}

func (*Field) ObjectType() string { return "astral.blueprint.field" }

func (f Field) WriteTo(w io.Writer) (int64, error)   { return Objectify(&f).WriteTo(w) }
func (f *Field) ReadFrom(r io.Reader) (int64, error) { return Objectify(f).ReadFrom(r) }

// isASCII reports whether s contains only printable ASCII bytes (0x20..0x7E).
func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < 0x20 || s[i] > 0x7E {
			return false
		}
	}
	return true
}

// MaxBlueprintNameLen caps the per-name byte length of Blueprint.Type and Field.Name. The wire
// allows up to 65 535 (String16), but bounded names keep a malicious registration from pinning
// megabytes of permanent state in the registry per Blueprint. 255 matches String8's natural
// ceiling and leaves comfortable headroom over real identifiers (~40 chars in this codebase).
const MaxBlueprintNameLen = 255

// MaxArraySpecLength bounds ArraySpec.Length so a peer-registered schema can't allocate a
// multi-gigabyte array on every NewRuntimeObject / readField. 65 536 mirrors the spirit of
// MaxBlueprintNameLen — bounded enough to defeat malicious registrations, generous enough
// to cover realistic fixed-length arrays.
const MaxArraySpecLength = 1 << 16

// validateBlueprint enforces v1 structural rules. Shared name discipline runs first
// (non-empty ASCII Type ≤MaxBlueprintNameLen); the kind-specific body follows:
// alias kind validates Underlying against the primitive allowlist; struct kind validates
// each Field. Setting both Fields and Underlying is rejected as kind-ambiguous.
func validateBlueprint(bp *Blueprint) error {
	if bp == nil || bp.Type.String() == "" {
		return fmt.Errorf("%w: empty Type", ErrBlueprintInvalid)
	}
	// why: identifiers are ASCII across the codebase; restricting Type and Field.Name to ASCII
	// keeps ObjectIDs portable and eliminates the NFC/NFD divergence + surrogate hazards that
	// silently produce different hashes for visually identical names.
	// todo: tighten further to an identifier grammar (segment ("." segment)*, segment = [a-z]
	// [a-z0-9_]*) to catch typos like leading digits, spaces, or hyphens.
	if !isASCII(bp.Type.String()) {
		return fmt.Errorf("%w: Type must be ASCII", ErrBlueprintInvalid)
	}
	if len(bp.Type) > MaxBlueprintNameLen {
		return fmt.Errorf("%w: Type exceeds %d bytes", ErrBlueprintInvalid, MaxBlueprintNameLen)
	}
	if bp.Underlying.String() != "" && len(bp.Fields) > 0 {
		return fmt.Errorf("%w: kind ambiguous — both Fields and Underlying set", ErrBlueprintInvalid)
	}

	if bp.Kind() == BlueprintKindAlias {
		if !isAllowedPrimitive(bp.Underlying.String()) {
			return fmt.Errorf("%w: Underlying %q not in primitive allowlist",
				ErrBlueprintInvalid, bp.Underlying)
		}
		return nil
	}

	// why: dedupe case-insensitively so the binary-validated schema matches what the
	// JSON decoder accepts (RuntimeObject.UnmarshalJSON folds keys to lowercase). Without
	// this, a Blueprint with fields "Foo" and "FOO" registers and round-trips through the
	// binary codec but fails MarshalJSON→UnmarshalJSON with a duplicate-fold error.
	seen := map[string]bool{}
	for _, f := range bp.Fields {
		name := f.Name.String()
		if name == "" {
			return fmt.Errorf("%w: empty Field.Name", ErrBlueprintInvalid)
		}
		if !isASCII(name) {
			return fmt.Errorf("%w: Field.Name %q must be ASCII", ErrBlueprintInvalid, name)
		}
		if len(f.Name) > MaxBlueprintNameLen {
			return fmt.Errorf("%w: Field.Name exceeds %d bytes", ErrBlueprintInvalid, MaxBlueprintNameLen)
		}
		key := strings.ToLower(name)
		if seen[key] {
			return fmt.Errorf("%w: duplicate Field %s (case-insensitive)", ErrBlueprintInvalid, name)
		}
		seen[key] = true
		// why: RefSpec/PtrSpec to self produce unbounded decode recursion that consumes zero
		// (RefSpec) or one (PtrSpec presence) byte per frame — stack-overflow on a single
		// instance. SliceSpec/MapSpec/ArraySpec self-reference is bounded by the wire count
		// and stays within structural recursion limits, so we don't reject those here.
		switch sp := f.Spec.(type) {
		case *RefSpec:
			if sp.Type.String() == bp.Type.String() {
				return fmt.Errorf("%w: self-referential RefSpec %s", ErrBlueprintInvalid, sp.Type)
			}
		case *PtrSpec:
			if sp.Type.String() == bp.Type.String() {
				return fmt.Errorf("%w: self-referential PtrSpec %s", ErrBlueprintInvalid, sp.Type)
			}
		}
		if err := validateSpec(f.Spec); err != nil {
			return err
		}
	}
	return nil
}

func validateSpec(s Spec) error {
	switch v := s.(type) {
	case *PrimitiveSpec:
		if !isAllowedPrimitive(v.PrimitiveType.String()) {
			return fmt.Errorf("%w: primitive %q not in allowlist", ErrBlueprintInvalid, v.PrimitiveType)
		}
	case *RefSpec:
		if v.Type.String() == "" {
			return fmt.Errorf("%w: RefSpec.Type empty", ErrBlueprintInvalid)
		}

	case *SliceSpec:
		// note: empty Type allowed (heterogeneous slice)
	case *ArraySpec:
		// note: empty Type allowed (heterogeneous array); Length is part of the schema.
		// why: cap Length to defeat a peer-registered schema with Length≈4e9 forcing every
		// NewRuntimeObject / readField to allocate gigabytes via reflect.ArrayOf.
		if uint64(v.Length) > MaxArraySpecLength {
			return fmt.Errorf("%w: ArraySpec.Length %d exceeds %d",
				ErrBlueprintInvalid, v.Length, MaxArraySpecLength)
		}
	case *MapSpec:
		if !isAllowedMapKey(v.KeyType.String()) {
			return fmt.Errorf("%w: MapSpec.KeyType %q not in allowlist", ErrBlueprintInvalid, v.KeyType)
		}
	case *PtrSpec:
		if v.Type.String() == "" {
			return fmt.Errorf("%w: PtrSpec.Type empty", ErrBlueprintInvalid)
		}
	case *ObjectSpec:
		// no fields
	default:
		return fmt.Errorf("%w: unknown Spec %T", ErrBlueprintInvalid, s)
	}
	return nil
}

// PrimitiveAlias lets a compile-time Object prototype declare itself as a primitive alias so
// the registry can derive an alias-kind Blueprint for it on the fly — the Go-side
// counterpart of `type Foo astral.Uint8`. Apps add a single method:
//
//	func (*Mode) UnderlyingPrimitive() string { return "uint8" }
//
// `astral.Add(&Mode{})` is then enough: local New() returns the typed Go value, and
// BlueprintOf derives an alias-kind Blueprint for sync without an explicit RegisterBlueprint.
type PrimitiveAlias interface {
	UnderlyingPrimitive() string
}

func init() {
	_ = Add(&Blueprint{})
	_ = Add(&Field{})
}
