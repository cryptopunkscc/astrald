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
// polymorphically (type tag + payload) by Encode/Decode.
type Field struct {
	Name String16
	Spec Object
}

func (*Field) ObjectType() string { return "astral.blueprint.field" }

func (f Field) WriteTo(w io.Writer) (int64, error)   { return Objectify(&f).WriteTo(w) }
func (f *Field) ReadFrom(r io.Reader) (int64, error) { return Objectify(f).ReadFrom(r) }

// validateBlueprint enforces v1 structural rules: non-empty Type, non-nil Fields, unique
// field names, and each Spec drawn from the known carriers within its allowlist.
func validateBlueprint(bp *Blueprint) error {
	if bp == nil || bp.Type.String() == "" {
		return fmt.Errorf("%w: empty Type", ErrBlueprintInvalid)
	}

	seen := map[string]bool{}
	for _, f := range bp.Fields {
		name := f.Name.String()
		if name == "" {
			return fmt.Errorf("%w: empty Field.PrimitiveType", ErrBlueprintInvalid)
		}
		if seen[name] {
			return fmt.Errorf("%w: duplicate Field %s", ErrBlueprintInvalid, name)
		}
		seen[name] = true
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
