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

func TestOrderedTypes_CompileTimeOnly_AlphaSorted(t *testing.T) {
	bps := NewBlueprints(nil)
	_ = bps.Add(&Blueprint{}, &Field{}) // ObjectType: astral.blueprint, astral.blueprint.field

	got := bps.OrderedBlueprints()
	want := []string{"astral.blueprint", "astral.blueprint.field"}
	if !equalStrings(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestOrderedTypes_RuntimeAfterCompileTime(t *testing.T) {
	bps := NewBlueprints(nil)
	bp := NewBlueprint("test.runtime_one",
		Field{Name: "x", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
	)
	if _, err := bps.RegisterBlueprint(bp); err != nil {
		t.Fatal(err)
	}

	got := bps.OrderedBlueprints()
	if len(got) == 0 || got[len(got)-1] != "test.runtime_one" {
		t.Fatalf("runtime entry must trail compile-time block, got %v", got)
	}
}

func TestOrderedTypes_RuntimeTopoByRef(t *testing.T) {
	bps := NewBlueprints(nil)

	leaf := NewBlueprint("test.leaf",
		Field{Name: "n", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
	)
	if _, err := bps.RegisterBlueprint(leaf); err != nil {
		t.Fatal(err)
	}
	mid := NewBlueprint("test.mid",
		Field{Name: "r", Spec: &RefSpec{Type: "test.leaf"}},
	)
	if _, err := bps.RegisterBlueprint(mid); err != nil {
		t.Fatal(err)
	}
	top := NewBlueprint("test.top",
		Field{Name: "r", Spec: &RefSpec{Type: "test.mid"}},
	)
	if _, err := bps.RegisterBlueprint(top); err != nil {
		t.Fatal(err)
	}

	got := bps.OrderedBlueprints()
	mustPrecede(t, got, "test.leaf", "test.mid")
	mustPrecede(t, got, "test.mid", "test.top")
}

func TestOrderedTypes_ParentChainPrecedesChild(t *testing.T) {
	parent := NewBlueprints(nil)
	if _, err := parent.RegisterBlueprint(NewBlueprint("test.parent_only",
		Field{Name: "n", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
	)); err != nil {
		t.Fatal(err)
	}

	child := NewBlueprints(parent)
	if _, err := child.RegisterBlueprint(NewBlueprint("test.child_only",
		Field{Name: "r", Spec: &RefSpec{Type: "test.parent_only"}},
	)); err != nil {
		t.Fatal(err)
	}

	got := child.OrderedBlueprints()
	mustPrecede(t, got, "test.parent_only", "test.child_only")
}

func TestOrderedTypes_StableTieBreakAlpha(t *testing.T) {
	bps := NewBlueprints(nil)
	names := []string{"test.b_root", "test.a_root", "test.c_root"}
	for _, n := range names {
		if _, err := bps.RegisterBlueprint(NewBlueprint(n,
			Field{Name: "n", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
		)); err != nil {
			t.Fatal(err)
		}
	}

	got := bps.OrderedBlueprints()
	idx := map[string]int{}
	for i, n := range got {
		idx[n] = i
	}
	if !(idx["test.a_root"] < idx["test.b_root"] && idx["test.b_root"] < idx["test.c_root"]) {
		t.Fatalf("expected alpha tie-break a<b<c, got %v", got)
	}
}

func mustPrecede(t *testing.T, names []string, earlier, later string) {
	t.Helper()
	ei, li := -1, -1
	for i, n := range names {
		if n == earlier {
			ei = i
		}
		if n == later {
			li = i
		}
	}
	if ei < 0 {
		t.Fatalf("%q missing from %v", earlier, names)
	}
	if li < 0 {
		t.Fatalf("%q missing from %v", later, names)
	}
	if ei >= li {
		t.Fatalf("expected %q before %q, got %v", earlier, later, names)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestAllBlueprints_CompileTimeReflected(t *testing.T) {
	bps := NewBlueprints(nil)
	_ = bps.Add(&Blueprint{}, &Field{})

	got, err := bps.AllBlueprints()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 blueprints, got %d", len(got))
	}
	if got[0].Type.String() != "astral.blueprint" || got[1].Type.String() != "astral.blueprint.field" {
		t.Fatalf("want compile-time alpha order, got %s,%s",
			got[0].Type, got[1].Type)
	}
	if len(got[0].Fields) == 0 {
		t.Fatalf("derived Blueprint must carry reflected fields, got empty")
	}
}

func TestAllBlueprints_RuntimeTopoAfterCompileTime(t *testing.T) {
	bps := NewBlueprints(nil)
	_ = bps.Add(&Blueprint{})

	leaf := NewBlueprint("test.all_leaf",
		Field{Name: "n", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
	)
	if _, err := bps.RegisterBlueprint(leaf); err != nil {
		t.Fatal(err)
	}
	top := NewBlueprint("test.all_top",
		Field{Name: "r", Spec: &RefSpec{Type: "test.all_leaf"}},
	)
	if _, err := bps.RegisterBlueprint(top); err != nil {
		t.Fatal(err)
	}

	got, err := bps.AllBlueprints()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	names := make([]string, len(got))
	for i, b := range got {
		names[i] = b.Type.String()
	}
	if names[0] != "astral.blueprint" {
		t.Fatalf("compile-time must come first, got %v", names)
	}
	mustPrecede(t, names, "test.all_leaf", "test.all_top")
}

func TestAllBlueprints_BadPrototypeAggregated(t *testing.T) {
	bps := NewBlueprints(nil)
	b := Bool(false)
	_ = bps.Add(&Blueprint{}, &b)

	got, err := bps.AllBlueprints()
	if err == nil {
		t.Fatalf("expected aggregated error for non-struct prototype")
	}
	if len(got) != 1 || got[0].Type.String() != "astral.blueprint" {
		t.Fatalf("want only the good entry, got %d entries", len(got))
	}
}

func TestAllBlueprints_ParentChain(t *testing.T) {
	parent := NewBlueprints(nil)
	_ = parent.Add(&Blueprint{})

	child := NewBlueprints(parent)
	if _, err := child.RegisterBlueprint(NewBlueprint("test.all_child",
		Field{Name: "n", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
	)); err != nil {
		t.Fatal(err)
	}

	got, err := child.AllBlueprints()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	names := make([]string, len(got))
	for i, b := range got {
		names[i] = b.Type.String()
	}
	mustPrecede(t, names, "astral.blueprint", "test.all_child")
}
