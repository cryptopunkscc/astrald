package astral

import (
	"fmt"
	"io"
)

var _ Object = (*Blueprint)(nil)

// Blueprint describes a runtime-registered Object type by name and an ordered list of fields.
type Blueprint struct {
	Type   String16
	Fields []Field
}

func (*Blueprint) ObjectType() string { return "astral.blueprint" }

// NewBlueprint returns a Blueprint with the given type name and fields.
func NewBlueprint(typeName string, fields ...Field) *Blueprint {
	return &Blueprint{Type: String16(typeName), Fields: append([]Field{}, fields...)}
}

func (bp Blueprint) WriteTo(w io.Writer) (int64, error)   { return Objectify(&bp).WriteTo(w) }
func (bp *Blueprint) ReadFrom(r io.Reader) (int64, error) { return Objectify(bp).ReadFrom(r) }

var _ Object = (*Field)(nil)

// Field is a single named slot inside a Blueprint. Spec is one of the Spec carriers
// (*PrimitiveSpec, *RefSpec, *SliceSpec, *MapSpec, *PtrSpec, *ObjectSpec) and is encoded
// polymorphically (type tag + payload) by Encode/Decode. Spec wire bytes are tag-less on
// their own; the discriminator comes from this interface-typed slot.
type Field struct {
	Name String16
	Spec Object
}

func (*Field) ObjectType() string { return "astral.blueprint.field" }

func (f Field) WriteTo(w io.Writer) (int64, error)   { return Objectify(&f).WriteTo(w) }
func (f *Field) ReadFrom(r io.Reader) (int64, error) { return Objectify(f).ReadFrom(r) }

// isASCII reports whether s contains only bytes in the 7-bit ASCII range.
func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
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

// validateBlueprint enforces v1 structural rules: non-empty Type, non-nil Fields, unique
// field names, and each Spec drawn from the known carriers within its allowlist.
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
		if seen[name] {
			return fmt.Errorf("%w: duplicate Field %s", ErrBlueprintInvalid, name)
		}
		seen[name] = true
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

func validateSpec(s Object) error {
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

func init() {
	_ = Add(&Blueprint{})
	_ = Add(&Field{})
}
