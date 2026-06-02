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

func TestRuntimeObject_ArrayRoundTrip(t *testing.T) {
	bp := NewBlueprint("test.array",
		Field{Name: "items", Spec: &ArraySpec{Type: "uint32", Length: 3}},
	)

	src := NewRuntimeObject(bp)
	ra, err := newRuntimeArray("uint32", 3)
	if err != nil {
		t.Fatal(err)
	}
	for i, v := range []uint32{10, 20, 30} {
		if err := ra.Set(i, NewUint32(v)); err != nil {
			t.Fatal(err)
		}
	}
	if err := src.Set("items", ra); err != nil {
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

	got, ok := dst.Get("items").(*runtimeArray)
	if !ok {
		t.Fatalf("items: want *runtimeArray, got %T", dst.Get("items"))
	}
	if got.Len() != 3 {
		t.Fatalf("len: want 3, got %d", got.Len())
	}
	for i, want := range []uint32{10, 20, 30} {
		u, ok := got.At(i).(*Uint32)
		if !ok || *u != Uint32(want) {
			t.Fatalf("[%d]: want %d, got %#v", i, want, got.At(i))
		}
	}
}

func TestRuntimeObject_ArrayHeterogeneousRoundTrip(t *testing.T) {
	bp := NewBlueprint("test.array.hetero",
		Field{Name: "mixed", Spec: &ArraySpec{Type: "", Length: 2}},
	)

	src := NewRuntimeObject(bp)
	ra, err := newRuntimeArray("", 2)
	if err != nil {
		t.Fatal(err)
	}
	if err := ra.Set(0, NewUint32(7)); err != nil {
		t.Fatal(err)
	}
	if err := ra.Set(1, NewString16("hi")); err != nil {
		t.Fatal(err)
	}
	if err := src.Set("mixed", ra); err != nil {
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

	got, _ := dst.Get("mixed").(*runtimeArray)
	if got == nil || got.Len() != 2 {
		t.Fatalf("mixed: want len 2, got %#v", dst.Get("mixed"))
	}
	if u, ok := got.At(0).(*Uint32); !ok || *u != 7 {
		t.Fatalf("[0]: want 7, got %#v", got.At(0))
	}
	if s, ok := got.At(1).(*String16); !ok || *s != "hi" {
		t.Fatalf("[1]: want \"hi\", got %#v", got.At(1))
	}
}

func TestRuntimeObject_UnregisteredArrayElement_FailsSymmetrically(t *testing.T) {
	bp := NewBlueprint("test.array.missing",
		Field{Name: "a", Spec: &ArraySpec{Type: "definitely-not-registered", Length: 2}},
	)

	src := NewRuntimeObject(bp)
	var buf bytes.Buffer
	_, err := src.WriteTo(&buf)
	if !errors.Is(err, ErrBlueprintNotFound) {
		t.Fatalf("encode: want ErrBlueprintNotFound, got %v", err)
	}

	dst := NewRuntimeObject(bp)
	_, err = dst.ReadFrom(&buf)
	if !errors.Is(err, ErrBlueprintNotFound) {
		t.Fatalf("decode: want ErrBlueprintNotFound, got %v", err)
	}
}

func TestRuntimeObject_ArrayLengthMismatch(t *testing.T) {
	bp := NewBlueprint("test.array.len",
		Field{Name: "items", Spec: &ArraySpec{Type: "uint32", Length: 3}},
	)
	ro := NewRuntimeObject(bp)
	ra, err := newRuntimeArray("uint32", 2) // wrong length
	if err != nil {
		t.Fatal(err)
	}
	if err := ro.Set("items", ra); !errors.Is(err, ErrFieldTypeMismatch) {
		t.Fatalf("want ErrFieldTypeMismatch, got %v", err)
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
	if _, ok := dstNil.Get("p").(*Nil); !ok {
		t.Fatalf("expected *Nil, got %#v", dstNil.Get("p"))
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
	if _, ok := ro.Get("opt").(*Nil); !ok {
		t.Fatalf("opt: want *Nil, got %#v", ro.Get("opt"))
	}
}

// why: pins codec symmetry for SliceSpec/MapSpec whose element type is unregistered. The
// previous specZero silently substituted a heterogeneous carrier, so WriteTo emitted
// per-element-tagged bytes while ReadFrom errored — same Blueprint, divergent wire shape.
func TestRuntimeObject_UnregisteredSliceElement_FailsSymmetrically(t *testing.T) {
	bp := NewBlueprint("test.slice.missing",
		Field{Name: "items", Spec: &SliceSpec{Type: "definitely-not-registered"}},
	)

	src := NewRuntimeObject(bp)
	var buf bytes.Buffer
	_, err := src.WriteTo(&buf)
	if !errors.Is(err, ErrBlueprintNotFound) {
		t.Fatalf("encode: want ErrBlueprintNotFound, got %v", err)
	}

	dst := NewRuntimeObject(bp)
	_, err = dst.ReadFrom(&buf)
	if !errors.Is(err, ErrBlueprintNotFound) {
		t.Fatalf("decode: want ErrBlueprintNotFound, got %v", err)
	}
}

func TestRuntimeObject_UnregisteredMapValue_FailsSymmetrically(t *testing.T) {
	bp := NewBlueprint("test.map.missing",
		Field{Name: "m", Spec: &MapSpec{KeyType: "string16", ValueType: "definitely-not-registered"}},
	)

	src := NewRuntimeObject(bp)
	var buf bytes.Buffer
	_, err := src.WriteTo(&buf)
	if !errors.Is(err, ErrBlueprintNotFound) {
		t.Fatalf("encode: want ErrBlueprintNotFound, got %v", err)
	}

	dst := NewRuntimeObject(bp)
	_, err = dst.ReadFrom(&buf)
	if !errors.Is(err, ErrBlueprintNotFound) {
		t.Fatalf("decode: want ErrBlueprintNotFound, got %v", err)
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

// TestRuntimeObject_DecodeDepthCap_MutualPtr proves the decode depth cap converts
// a mutually-recursive Blueprint pair (X.PtrSpec→Y, Y.PtrSpec→X) into a typed
// ErrDepthExceeded instead of a stack overflow when fed a stream that keeps every
// presence byte non-zero. Registration-time cycle detection is a separate change;
// this test exercises the decode-side safety net.
func TestRuntimeObject_DecodeDepthCap_MutualPtr(t *testing.T) {
	xBP := NewBlueprint("test.depth.x",
		Field{Name: "y", Spec: &PtrSpec{Type: "test.depth.y"}},
	)
	yBP := NewBlueprint("test.depth.y",
		Field{Name: "x", Spec: &PtrSpec{Type: "test.depth.x"}},
	)
	if _, err := RegisterBlueprint(xBP); err != nil {
		t.Fatalf("register X: %v", err)
	}
	if _, err := RegisterBlueprint(yBP); err != nil {
		t.Fatalf("register Y: %v", err)
	}

	ro := NewRuntimeObject(xBP)
	// All-ones stream: every PtrSpec presence byte decodes as "present", forcing the
	// decoder to recurse into the referenced type indefinitely (until the depth cap).
	r := bytes.NewReader(bytes.Repeat([]byte{0xFF}, 4*MaxBlueprintDepth))
	_, err := ro.ReadFrom(r)
	if !errors.Is(err, ErrDepthExceeded) {
		t.Fatalf("want ErrDepthExceeded, got %v", err)
	}
}
