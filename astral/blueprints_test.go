package astral

import (
	"errors"
	"strings"
	"testing"
)

func TestRegisterBlueprint_Success(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())

	bp := NewBlueprint("test.simple",
		Field{Name: "n", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
	)

	id, err := bps.RegisterBlueprint(bp)
	if err != nil {
		t.Fatal(err)
	}
	if id == nil {
		t.Fatal("expected non-nil ObjectID")
	}

	if got := bps.GetBlueprint("test.simple"); got == nil {
		t.Fatal("GetBlueprint returned nil for registered Type")
	}
}

func TestRegisterBlueprint_EmptyType(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	bp := NewBlueprint("",
		Field{Name: "x", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
	)

	_, err := bps.RegisterBlueprint(bp)
	if !errors.Is(err, ErrBlueprintInvalid) {
		t.Fatalf("want ErrBlueprintInvalid, got %v", err)
	}
}

func TestRegisterBlueprint_DuplicateName(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	bp := NewBlueprint("dup.case",
		Field{Name: "x", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
	)

	if _, err := bps.RegisterBlueprint(bp); err != nil {
		t.Fatal(err)
	}

	// re-register same type
	bp2 := NewBlueprint("dup.case",
		Field{Name: "y", Spec: &PrimitiveSpec{PrimitiveType: "uint64"}},
	)
	_, err := bps.RegisterBlueprint(bp2)
	if !errors.Is(err, ErrAlreadyRegistered) {
		t.Fatalf("want ErrAlreadyRegistered, got %v", err)
	}
}

func TestRegisterBlueprint_CollidesWithPrototype(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	bp := NewBlueprint("astral.blueprint",
		Field{Name: "x", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
	)

	_, err := bps.RegisterBlueprint(bp)
	if !errors.Is(err, ErrAlreadyRegistered) {
		t.Fatalf("want ErrAlreadyRegistered, got %v", err)
	}
}

func TestRegisterBlueprint_BadPrimitive(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	bp := NewBlueprint("test.bad_primitive",
		Field{Name: "x", Spec: &PrimitiveSpec{PrimitiveType: "int32"}},
	)

	_, err := bps.RegisterBlueprint(bp)
	if !errors.Is(err, ErrBlueprintInvalid) {
		t.Fatalf("want ErrBlueprintInvalid, got %v", err)
	}
}

func TestRegisterBlueprint_BadMapKey(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	bp := NewBlueprint("test.bad_map_key",
		Field{Name: "m", Spec: &MapSpec{KeyType: "bytes16", ValueType: "uint32"}},
	)

	_, err := bps.RegisterBlueprint(bp)
	if !errors.Is(err, ErrBlueprintInvalid) {
		t.Fatalf("want ErrBlueprintInvalid, got %v", err)
	}
}

func TestRegisterBlueprint_MapKeyAllowlistSweep(t *testing.T) {
	allowed := map[string]bool{
		"string16": true,
		"uint8":    true,
		"uint16":   true,
		"uint32":   true,
		"uint64":   true,
	}
	// Cover every primitive plus a few non-primitives to lock the contract.
	candidates := []string{
		"string8", "string16", "string32", "string64",
		"uint8", "uint16", "uint32", "uint64",
		"bytes8", "bytes16", "bytes32", "bytes64",
		"bool", "time", "duration", "identity",
		"int32",
	}
	for _, name := range candidates {
		t.Run(name, func(t *testing.T) {
			bps := NewBlueprints(DefaultBlueprints())
			bp := NewBlueprint("test.mapkey."+name,
				Field{Name: "m", Spec: &MapSpec{KeyType: String16(name), ValueType: "uint32"}},
			)
			_, err := bps.RegisterBlueprint(bp)
			if allowed[name] {
				if err != nil {
					t.Fatalf("key %q should be allowed, got %v", name, err)
				}
			} else {
				if !errors.Is(err, ErrBlueprintInvalid) {
					t.Fatalf("key %q should be rejected, got %v", name, err)
				}
			}
		})
	}
}

func TestRegisterBlueprint_EmptyRefType(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	bp := NewBlueprint("test.empty_ref",
		Field{Name: "r", Spec: &RefSpec{Type: ""}},
	)

	_, err := bps.RegisterBlueprint(bp)
	if !errors.Is(err, ErrBlueprintInvalid) {
		t.Fatalf("want ErrBlueprintInvalid, got %v", err)
	}
}

func TestRegisterBlueprint_RejectsOversizedType(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	longType := strings.Repeat("a", MaxBlueprintNameLen+1)
	bp := NewBlueprint(longType,
		Field{Name: "x", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
	)
	_, err := bps.RegisterBlueprint(bp)
	if !errors.Is(err, ErrBlueprintInvalid) {
		t.Fatalf("want ErrBlueprintInvalid, got %v", err)
	}
}

func TestRegisterBlueprint_RejectsOversizedFieldName(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	longName := strings.Repeat("a", MaxBlueprintNameLen+1)
	bp := NewBlueprint("test.oversized_field",
		Field{Name: String16(longName), Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
	)
	_, err := bps.RegisterBlueprint(bp)
	if !errors.Is(err, ErrBlueprintInvalid) {
		t.Fatalf("want ErrBlueprintInvalid, got %v", err)
	}
}

func TestRegisterBlueprint_DuplicateField(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	bp := NewBlueprint("test.dup_field",
		Field{Name: "n", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
		Field{Name: "n", Spec: &PrimitiveSpec{PrimitiveType: "uint64"}},
	)

	_, err := bps.RegisterBlueprint(bp)
	if !errors.Is(err, ErrBlueprintInvalid) {
		t.Fatalf("want ErrBlueprintInvalid, got %v", err)
	}
}

func TestRegisterBlueprint_NonASCIIFieldName(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	bp := NewBlueprint("test.nonascii_field",
		Field{Name: "café", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
	)
	_, err := bps.RegisterBlueprint(bp)
	if !errors.Is(err, ErrBlueprintInvalid) {
		t.Fatalf("want ErrBlueprintInvalid, got %v", err)
	}
}

func TestRegisterBlueprint_NonASCIIType(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	bp := NewBlueprint("test.café",
		Field{Name: "x", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
	)
	_, err := bps.RegisterBlueprint(bp)
	if !errors.Is(err, ErrBlueprintInvalid) {
		t.Fatalf("want ErrBlueprintInvalid, got %v", err)
	}
}

func TestRegisterBlueprint_TypeAtExactMaxLen(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	bp := NewBlueprint(strings.Repeat("a", MaxBlueprintNameLen),
		Field{Name: "x", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
	)
	if _, err := bps.RegisterBlueprint(bp); err != nil {
		t.Fatalf("Type of exactly MaxBlueprintNameLen bytes must register, got %v", err)
	}
}

func TestRegisterBlueprint_EmptyFieldNameInMiddle(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	bp := NewBlueprint("test.empty_middle",
		Field{Name: "a", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
		Field{Name: "", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
		Field{Name: "c", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
	)
	_, err := bps.RegisterBlueprint(bp)
	if !errors.Is(err, ErrBlueprintInvalid) {
		t.Fatalf("want ErrBlueprintInvalid, got %v", err)
	}
	if !strings.Contains(err.Error(), "empty Field.Name") {
		t.Fatalf("want error to name the empty field, got %v", err)
	}
}

func TestRegisterBlueprint_SelfReferentialRefSpec(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	bp := NewBlueprint("test.self_ref_ref",
		Field{Name: "r", Spec: &RefSpec{Type: "test.self_ref_ref"}},
	)
	_, err := bps.RegisterBlueprint(bp)
	if !errors.Is(err, ErrBlueprintInvalid) {
		t.Fatalf("want ErrBlueprintInvalid, got %v", err)
	}
}

func TestRegisterBlueprint_SelfReferentialPtrSpec(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	bp := NewBlueprint("test.self_ref_ptr",
		Field{Name: "p", Spec: &PtrSpec{Type: "test.self_ref_ptr"}},
	)
	_, err := bps.RegisterBlueprint(bp)
	if !errors.Is(err, ErrBlueprintInvalid) {
		t.Fatalf("want ErrBlueprintInvalid, got %v", err)
	}
}

func TestRegisterBlueprint_BadPrimitiveSweep(t *testing.T) {
	bad := []string{"int32", "uint128", "byte", "float128", "string"}
	for _, name := range bad {
		t.Run(name, func(t *testing.T) {
			bps := NewBlueprints(DefaultBlueprints())
			bp := NewBlueprint("test.bad_primitive_"+name,
				Field{Name: "x", Spec: &PrimitiveSpec{PrimitiveType: String16(name)}},
			)
			_, err := bps.RegisterBlueprint(bp)
			if !errors.Is(err, ErrBlueprintInvalid) {
				t.Fatalf("want ErrBlueprintInvalid for %q, got %v", name, err)
			}
		})
	}
}

func TestRegisterBlueprint_ArraySpecZeroLength(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	bp := NewBlueprint("test.array_zero",
		Field{Name: "a", Spec: &ArraySpec{Type: "uint32", Length: 0}},
	)
	if _, err := bps.RegisterBlueprint(bp); err != nil {
		t.Fatalf("zero-length ArraySpec must register, got %v", err)
	}
}

func TestRegisterBlueprint_EmptyPtrType(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	bp := NewBlueprint("test.empty_ptr",
		Field{Name: "p", Spec: &PtrSpec{Type: ""}},
	)
	_, err := bps.RegisterBlueprint(bp)
	if !errors.Is(err, ErrBlueprintInvalid) {
		t.Fatalf("want ErrBlueprintInvalid, got %v", err)
	}
}

func TestGetBlueprint_CompileTimeReturnsNil(t *testing.T) {
	// compile-time prototype lives under "astral.blueprint" but its Type is empty, so
	// GetBlueprint must NOT return it.
	if got := DefaultBlueprints().GetBlueprint("astral.blueprint"); got != nil {
		t.Fatalf("expected nil for compile-time prototype, got %#v", got)
	}
}

func TestNew_ReturnsRuntimeObject(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	bp := NewBlueprint("test.runtime_new",
		Field{Name: "n", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
	)
	if _, err := bps.RegisterBlueprint(bp); err != nil {
		t.Fatal(err)
	}

	obj := bps.New("test.runtime_new")
	if obj == nil {
		t.Fatal("New returned nil for registered runtime Blueprint")
	}
	if _, ok := obj.(*RuntimeObject); !ok {
		t.Fatalf("expected *RuntimeObject, got %T", obj)
	}
	if obj.ObjectType() != "test.runtime_new" {
		t.Fatalf("want type test.runtime_new, got %s", obj.ObjectType())
	}
}

// TestRegisterBlueprint_UnresolvedRef pins the closure invariant: a RefSpec to an
// unregistered type must be rejected at registration so the codec path never sees a
// Blueprint with a dangling reference.
func TestRegisterBlueprint_UnresolvedRef(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	bp := NewBlueprint("test.unresolved_ref",
		Field{Name: "r", Spec: &RefSpec{Type: "never.registered.type"}},
	)
	_, err := bps.RegisterBlueprint(bp)
	if !errors.Is(err, ErrBlueprintInvalid) {
		t.Fatalf("want ErrBlueprintInvalid, got %v", err)
	}
}

func TestRegisterBlueprint_UnresolvedPtr(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	bp := NewBlueprint("test.unresolved_ptr",
		Field{Name: "p", Spec: &PtrSpec{Type: "never.registered.type"}},
	)
	_, err := bps.RegisterBlueprint(bp)
	if !errors.Is(err, ErrBlueprintInvalid) {
		t.Fatalf("want ErrBlueprintInvalid, got %v", err)
	}
}

func TestRegisterBlueprint_UnresolvedSliceElement(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	bp := NewBlueprint("test.unresolved_slice",
		Field{Name: "s", Spec: &SliceSpec{Type: "never.registered.type"}},
	)
	_, err := bps.RegisterBlueprint(bp)
	if !errors.Is(err, ErrBlueprintInvalid) {
		t.Fatalf("want ErrBlueprintInvalid, got %v", err)
	}
}

func TestRegisterBlueprint_UnresolvedArrayElement(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	bp := NewBlueprint("test.unresolved_array",
		Field{Name: "a", Spec: &ArraySpec{Type: "never.registered.type", Length: 2}},
	)
	_, err := bps.RegisterBlueprint(bp)
	if !errors.Is(err, ErrBlueprintInvalid) {
		t.Fatalf("want ErrBlueprintInvalid, got %v", err)
	}
}

func TestRegisterBlueprint_UnresolvedMapValue(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	bp := NewBlueprint("test.unresolved_map",
		Field{Name: "m", Spec: &MapSpec{KeyType: "string16", ValueType: "never.registered.type"}},
	)
	_, err := bps.RegisterBlueprint(bp)
	if !errors.Is(err, ErrBlueprintInvalid) {
		t.Fatalf("want ErrBlueprintInvalid, got %v", err)
	}
}

func TestRegisterBlueprint_ResolvedRefSucceeds(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	a := NewBlueprint("test.resolved_a",
		Field{Name: "x", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
	)
	if _, err := bps.RegisterBlueprint(a); err != nil {
		t.Fatal(err)
	}
	b := NewBlueprint("test.resolved_b",
		Field{Name: "r", Spec: &RefSpec{Type: "test.resolved_a"}},
	)
	if _, err := bps.RegisterBlueprint(b); err != nil {
		t.Fatalf("ref to registered type should succeed, got %v", err)
	}
}

// TestRegisterBlueprint_HeterogeneousContainers pins that empty element/value Type on
// Slice/Array/Map and ObjectSpec are intentionally open and bypass the closure check.
func TestRegisterBlueprint_HeterogeneousContainers(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	bp := NewBlueprint("test.heterogeneous",
		Field{Name: "s", Spec: &SliceSpec{Type: ""}},
		Field{Name: "a", Spec: &ArraySpec{Type: "", Length: 2}},
		Field{Name: "m", Spec: &MapSpec{KeyType: "string16", ValueType: ""}},
		Field{Name: "o", Spec: &ObjectSpec{}},
	)
	if _, err := bps.RegisterBlueprint(bp); err != nil {
		t.Fatalf("heterogeneous containers should register, got %v", err)
	}
}

func TestNew_CompileTimeStillReturnsPrototype(t *testing.T) {
	obj := DefaultBlueprints().New("astral.blueprint")
	if obj == nil {
		t.Fatal("astral.blueprint prototype lookup returned nil")
	}
	if _, ok := obj.(*Blueprint); !ok {
		t.Fatalf("expected *Blueprint, got %T", obj)
	}
}
