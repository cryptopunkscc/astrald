package astral

import (
	"bytes"
	"testing"
)

func TestSpecCarriers_RoundTrip(t *testing.T) {
	cases := []Object{
		&PrimitiveSpec{PrimitiveType: "uint32"},
		&RefSpec{Type: "some.thing"},
		&SliceSpec{Type: ""},
		&MapSpec{KeyType: "string16", ValueType: "uint32"},
		&PtrSpec{Type: "some.thing"},
		&ObjectSpec{},
	}

	for _, src := range cases {
		t.Run(src.ObjectType(), func(t *testing.T) {
			var buf bytes.Buffer
			if _, err := Encode(&buf, src); err != nil {
				t.Fatal(err)
			}

			got, _, err := Decode(&buf)
			if err != nil {
				t.Fatal(err)
			}
			if got.ObjectType() != src.ObjectType() {
				t.Fatalf("type mismatch: want %s got %s", src.ObjectType(), got.ObjectType())
			}
		})
	}
}

func TestBlueprint_RoundTrip(t *testing.T) {
	src := NewBlueprint("my.type",
		Field{Name: "name", Spec: &PrimitiveSpec{PrimitiveType: "string16"}},
		Field{Name: "count", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
		Field{Name: "extra", Spec: &ObjectSpec{}},
	)

	var buf bytes.Buffer
	if _, err := src.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	dst := &Blueprint{}
	if _, err := dst.ReadFrom(&buf); err != nil {
		t.Fatal(err)
	}

	if dst.Type.String() != "my.type" {
		t.Fatalf("want my.type, got %s", dst.Type)
	}
	if len(dst.Fields) != 3 {
		t.Fatalf("expected 3 fields, got %#v", dst.Fields)
	}

	for i, want := range []string{"name", "count", "extra"} {
		got := dst.Fields[i]
		if got.Name.String() != want {
			t.Fatalf("field[%d] name: want %s got %s", i, want, got.Name)
		}
	}
}

func TestBlueprint_EmptyFields(t *testing.T) {
	src := NewBlueprint("empty")

	var buf bytes.Buffer
	if _, err := src.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	dst := &Blueprint{}
	if _, err := dst.ReadFrom(&buf); err != nil {
		t.Fatal(err)
	}

	if dst.Type.String() != "empty" {
		t.Fatalf("want empty, got %s", dst.Type)
	}
	if len(dst.Fields) != 0 {
		t.Fatalf("expected 0 fields, got %#v", dst.Fields)
	}
}

func TestBlueprint_RegistryLookup(t *testing.T) {
	obj := New("astral.blueprint")
	if obj == nil {
		t.Fatal("astral.blueprint not registered")
	}
	if _, ok := obj.(*Blueprint); !ok {
		t.Fatalf("expected *Blueprint, got %T", obj)
	}

	for _, name := range []string{
		"astral.blueprint",
		"astral.blueprint.field",
		"astral.blueprint.primitive_spec",
		"astral.blueprint.ref_spec",
		"astral.blueprint.slice_spec",
		"astral.blueprint.map_spec",
		"astral.blueprint.ptr_spec",
		"astral.blueprint.object_spec",
	} {
		if New(name) == nil {
			t.Errorf("%s not registered", name)
		}
	}
}

func TestBlueprint_ContentAddressed(t *testing.T) {
	a := NewBlueprint("same",
		Field{Name: "x", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
	)
	b := NewBlueprint("same",
		Field{Name: "x", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
	)

	idA, err := ResolveObjectID(a)
	if err != nil {
		t.Fatal(err)
	}
	idB, err := ResolveObjectID(b)
	if err != nil {
		t.Fatal(err)
	}
	if idA.String() != idB.String() {
		t.Fatalf("expected identical IDs, got %s vs %s", idA, idB)
	}
}
