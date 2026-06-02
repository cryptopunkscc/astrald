package astral

import (
	"errors"
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

func TestNew_CompileTimeStillReturnsPrototype(t *testing.T) {
	obj := DefaultBlueprints().New("astral.blueprint")
	if obj == nil {
		t.Fatal("astral.blueprint prototype lookup returned nil")
	}
	if _, ok := obj.(*Blueprint); !ok {
		t.Fatalf("expected *Blueprint, got %T", obj)
	}
}
