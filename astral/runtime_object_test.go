package astral

import (
	"bytes"
	"errors"
	"testing"
)

func TestRuntimeObject_PrimitiveRoundTrip(t *testing.T) {
	bp := NewBlueprint("test.prim",
		Field{Name: "n", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
		Field{Name: "s", Spec: &PrimitiveSpec{PrimitiveType: "string16"}},
		Field{Name: "b", Spec: &PrimitiveSpec{PrimitiveType: "bool"}},
	)

	src := NewRuntimeObject(bp)
	if err := src.Set("n", uint32(42)); err != nil {
		t.Fatal(err)
	}
	if err := src.Set("s", "hello"); err != nil {
		t.Fatal(err)
	}
	if err := src.Set("b", true); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if _, err := src.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	dst := NewRuntimeObject(bp)
	if _, err := dst.ReadFrom(&buf); err != nil {
		t.Fatal(err)
	}

	if u, ok := dst.Get("n").(*Uint32); !ok || *u != 42 {
		t.Fatalf("field n: want 42, got %#v", dst.Get("n"))
	}
	if s, ok := dst.Get("s").(*String16); !ok || *s != "hello" {
		t.Fatalf("field s: want \"hello\", got %#v", dst.Get("s"))
	}
	if b, ok := dst.Get("b").(*Bool); !ok || *b != true {
		t.Fatalf("field b: want true, got %#v", dst.Get("b"))
	}
}

func TestRuntimeObject_SliceRoundTrip(t *testing.T) {
	bp := NewBlueprint("test.slice",
		Field{Name: "items", Spec: &SliceSpec{Type: "uint32"}},
	)

	src := NewRuntimeObject(bp)
	rs, err := newRuntimeSlice("uint32")
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range []uint32{1, 2, 3} {
		if err := rs.Append(NewUint32(v)); err != nil {
			t.Fatal(err)
		}
	}
	if err := src.Set("items", rs); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if _, err := src.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	dst := NewRuntimeObject(bp)
	if _, err := dst.ReadFrom(&buf); err != nil {
		t.Fatal(err)
	}

	got, ok := dst.Get("items").(*runtimeSlice)
	if !ok {
		t.Fatalf("items: want *runtimeSlice, got %T", dst.Get("items"))
	}
	if got.Len() != 3 {
		t.Fatalf("len: want 3, got %d", got.Len())
	}
	for i, want := range []uint32{1, 2, 3} {
		u, ok := got.At(i).(*Uint32)
		if !ok || *u != Uint32(want) {
			t.Fatalf("[%d]: want %d, got %#v", i, want, got.At(i))
		}
	}
}

func TestRuntimeObject_MapRoundTrip(t *testing.T) {
	bp := NewBlueprint("test.maps",
		Field{Name: "m", Spec: &MapSpec{KeyType: "string16", ValueType: "uint32"}},
		Field{Name: "im", Spec: &MapSpec{KeyType: "uint16", ValueType: "uint32"}},
	)

	src := NewRuntimeObject(bp)
	sm, err := newRuntimeMap("string16", "uint32")
	if err != nil {
		t.Fatal(err)
	}
	if err := sm.Set("a", NewUint32(1)); err != nil {
		t.Fatal(err)
	}
	if err := src.Set("m", sm); err != nil {
		t.Fatal(err)
	}

	im, err := newRuntimeMap("uint16", "uint32")
	if err != nil {
		t.Fatal(err)
	}
	if err := im.Set(uint64(5), NewUint32(50)); err != nil {
		t.Fatal(err)
	}
	if err := src.Set("im", im); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if _, err := src.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	dst := NewRuntimeObject(bp)
	if _, err := dst.ReadFrom(&buf); err != nil {
		t.Fatal(err)
	}

	gotSM, ok := dst.Get("m").(*runtimeMap)
	if !ok || gotSM.Len() != 1 {
		t.Fatalf("string map: want len 1, got %#v", dst.Get("m"))
	}
	v, _ := gotSM.Get("a")
	if u, ok := v.(*Uint32); !ok || *u != 1 {
		t.Fatalf("string map[a]: want 1, got %#v", v)
	}

	gotIM, ok := dst.Get("im").(*runtimeMap)
	if !ok || gotIM.Len() != 1 {
		t.Fatalf("int map: want len 1, got %#v", dst.Get("im"))
	}
	iv, _ := gotIM.Get(uint64(5))
	if u, ok := iv.(*Uint32); !ok || *u != 50 {
		t.Fatalf("int map[5]: want 50, got %#v", iv)
	}
}

func TestRuntimeObject_PtrRoundTrip(t *testing.T) {
	bp := NewBlueprint("test.ptr",
		Field{Name: "p", Spec: &PtrSpec{Type: "uint32"}},
	)

	// nil case
	srcNil := NewRuntimeObject(bp)
	var nilBuf bytes.Buffer
	if _, err := srcNil.WriteTo(&nilBuf); err != nil {
		t.Fatal(err)
	}
	dstNil := NewRuntimeObject(bp)
	if _, err := dstNil.ReadFrom(&nilBuf); err != nil {
		t.Fatal(err)
	}
	if dstNil.Get("p") != nil {
		t.Fatalf("expected nil, got %#v", dstNil.Get("p"))
	}

	// non-nil case
	src := NewRuntimeObject(bp)
	if err := src.Set("p", NewUint32(7)); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, err := src.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	dst := NewRuntimeObject(bp)
	if _, err := dst.ReadFrom(&buf); err != nil {
		t.Fatal(err)
	}
	got, ok := dst.Get("p").(*Uint32)
	if !ok || *got != 7 {
		t.Fatalf("expected 7, got %#v", dst.Get("p"))
	}
}

func TestRuntimeObject_ObjectSpecRoundTrip(t *testing.T) {
	bp := NewBlueprint("test.obj",
		Field{Name: "any", Spec: &ObjectSpec{}},
	)

	src := NewRuntimeObject(bp)
	if err := src.Set("any", NewString16("hi")); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if _, err := src.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	dst := NewRuntimeObject(bp)
	if _, err := dst.ReadFrom(&buf); err != nil {
		t.Fatal(err)
	}

	got, ok := dst.Get("any").(*String16)
	if !ok || *got != "hi" {
		t.Fatalf("want \"hi\", got %#v", dst.Get("any"))
	}
}

func TestRuntimeObject_SetRejectsMismatch(t *testing.T) {
	bp := NewBlueprint("test.reject",
		Field{Name: "n", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
	)
	ro := NewRuntimeObject(bp)

	// width mismatch
	err := ro.Set("n", uint16(1))
	if !errors.Is(err, ErrFieldTypeMismatch) {
		t.Fatalf("want ErrFieldTypeMismatch, got %v", err)
	}

	// signed not allowed
	err = ro.Set("n", int32(1))
	if !errors.Is(err, ErrFieldTypeMismatch) {
		t.Fatalf("want ErrFieldTypeMismatch, got %v", err)
	}

	// unknown field
	err = ro.Set("nope", uint32(1))
	if !errors.Is(err, ErrFieldTypeMismatch) {
		t.Fatalf("want ErrFieldTypeMismatch, got %v", err)
	}
}

func TestRuntimeObject_NilOnlyForPtrSpec(t *testing.T) {
	bp := NewBlueprint("test.nil",
		Field{Name: "prim", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
		Field{Name: "opt", Spec: &PtrSpec{Type: "uint32"}},
	)
	ro := NewRuntimeObject(bp)

	if err := ro.Set("prim", nil); !errors.Is(err, ErrFieldTypeMismatch) {
		t.Fatalf("primitive accepted nil: %v", err)
	}
	if err := ro.Set("opt", nil); err != nil {
		t.Fatalf("PtrSpec rejected nil: %v", err)
	}
	if ro.Get("opt") != nil {
		t.Fatalf("opt: want nil, got %#v", ro.Get("opt"))
	}
}

func TestRuntimeObject_ContentAddressed(t *testing.T) {
	bp := NewBlueprint("test.id",
		Field{Name: "n", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
		Field{Name: "s", Spec: &PrimitiveSpec{PrimitiveType: "string16"}},
	)

	a := NewRuntimeObject(bp)
	_ = a.Set("n", uint32(7))
	_ = a.Set("s", "x")

	b := NewRuntimeObject(bp)
	_ = b.Set("n", uint32(7))
	_ = b.Set("s", "x")

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
