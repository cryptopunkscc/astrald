package astral

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func mustRuntimeObject(t *testing.T, bp *Blueprint) *RuntimeObject {
	t.Helper()
	ro, err := NewRuntimeObject(bp)
	if err != nil {
		t.Fatal(err)
	}
	return ro
}

func TestRuntimeObject_PrimitiveRoundTrip(t *testing.T) {
	bp := NewBlueprint("test.prim",
		Field{Name: "n", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
		Field{Name: "s", Spec: &PrimitiveSpec{PrimitiveType: "string16"}},
		Field{Name: "b", Spec: &PrimitiveSpec{PrimitiveType: "bool"}},
	)

	src := mustRuntimeObject(t, bp)
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

	dst := mustRuntimeObject(t, bp)
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

	src := mustRuntimeObject(t, bp)
	rs, err := NewRuntimeSlice("uint32")
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

	dst := mustRuntimeObject(t, bp)
	if _, err := dst.ReadFrom(&buf); err != nil {
		t.Fatal(err)
	}

	got, ok := dst.Get("items").(*RuntimeSlice)
	if !ok {
		t.Fatalf("items: want *RuntimeSlice, got %T", dst.Get("items"))
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

	src := mustRuntimeObject(t, bp)
	sm, err := NewRuntimeMap("string16", "uint32")
	if err != nil {
		t.Fatal(err)
	}
	if err := sm.Set("a", NewUint32(1)); err != nil {
		t.Fatal(err)
	}
	if err := src.Set("m", sm); err != nil {
		t.Fatal(err)
	}

	im, err := NewRuntimeMap("uint16", "uint32")
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

	dst := mustRuntimeObject(t, bp)
	if _, err := dst.ReadFrom(&buf); err != nil {
		t.Fatal(err)
	}

	gotSM, ok := dst.Get("m").(*RuntimeMap)
	if !ok || gotSM.Len() != 1 {
		t.Fatalf("string map: want len 1, got %#v", dst.Get("m"))
	}
	v, _ := gotSM.Get("a")
	if u, ok := v.(*Uint32); !ok || *u != 1 {
		t.Fatalf("string map[a]: want 1, got %#v", v)
	}

	gotIM, ok := dst.Get("im").(*RuntimeMap)
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

	src := mustRuntimeObject(t, bp)
	ra, err := NewRuntimeArray("uint32", 3)
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

	dst := mustRuntimeObject(t, bp)
	if _, err := dst.ReadFrom(&buf); err != nil {
		t.Fatal(err)
	}

	got, ok := dst.Get("items").(*RuntimeArray)
	if !ok {
		t.Fatalf("items: want *RuntimeArray, got %T", dst.Get("items"))
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

	src := mustRuntimeObject(t, bp)
	ra, err := NewRuntimeArray("", 2)
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

	dst := mustRuntimeObject(t, bp)
	if _, err := dst.ReadFrom(&buf); err != nil {
		t.Fatal(err)
	}

	got, _ := dst.Get("mixed").(*RuntimeArray)
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

	src := mustRuntimeObject(t, bp)
	var buf bytes.Buffer
	_, err := src.WriteTo(&buf)
	if !errors.Is(err, ErrBlueprintNotFound) {
		t.Fatalf("encode: want ErrBlueprintNotFound, got %v", err)
	}

	dst := mustRuntimeObject(t, bp)
	_, err = dst.ReadFrom(&buf)
	if !errors.Is(err, ErrBlueprintNotFound) {
		t.Fatalf("decode: want ErrBlueprintNotFound, got %v", err)
	}
}

func TestRuntimeObject_ArrayLengthMismatch(t *testing.T) {
	bp := NewBlueprint("test.array.len",
		Field{Name: "items", Spec: &ArraySpec{Type: "uint32", Length: 3}},
	)
	ro := mustRuntimeObject(t, bp)
	ra, err := NewRuntimeArray("uint32", 2) // wrong length
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
	srcNil := mustRuntimeObject(t, bp)
	var nilBuf bytes.Buffer
	if _, err := srcNil.WriteTo(&nilBuf); err != nil {
		t.Fatal(err)
	}
	dstNil := mustRuntimeObject(t, bp)
	if _, err := dstNil.ReadFrom(&nilBuf); err != nil {
		t.Fatal(err)
	}
	if _, ok := dstNil.Get("p").(*Nil); !ok {
		t.Fatalf("expected *Nil, got %#v", dstNil.Get("p"))
	}

	// non-nil case
	src := mustRuntimeObject(t, bp)
	if err := src.Set("p", NewUint32(7)); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, err := src.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	dst := mustRuntimeObject(t, bp)
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

	src := mustRuntimeObject(t, bp)
	if err := src.Set("any", NewString16("hi")); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if _, err := src.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	dst := mustRuntimeObject(t, bp)
	if _, err := dst.ReadFrom(&buf); err != nil {
		t.Fatal(err)
	}

	got, ok := dst.Get("any").(*String16)
	if !ok || *got != "hi" {
		t.Fatalf("want \"hi\", got %#v", dst.Get("any"))
	}
}

func TestRuntimeObject_ObjectSpec_NilRoundTrip(t *testing.T) {
	bp := NewBlueprint("test.obj.nil",
		Field{Name: "any", Spec: &ObjectSpec{}},
	)
	src := mustRuntimeObject(t, bp)
	// specZero installs &Nil{} for ObjectSpec; leave the field untouched.

	var buf bytes.Buffer
	if _, err := src.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	dst := mustRuntimeObject(t, bp)
	if _, err := dst.ReadFrom(&buf); err != nil {
		t.Fatal(err)
	}
	got := dst.Get("any")
	if _, ok := got.(*Nil); !ok {
		t.Fatalf("ObjectSpec nil round-trip: want *Nil, got %T", got)
	}
}

func TestRuntimeObject_SetUnknownField_NamesField(t *testing.T) {
	bp := NewBlueprint("test.set_unknown",
		Field{Name: "n", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
	)
	ro := mustRuntimeObject(t, bp)

	err := ro.Set("nope", uint32(1))
	if !errors.Is(err, ErrFieldTypeMismatch) {
		t.Fatalf("want ErrFieldTypeMismatch, got %v", err)
	}
	if !strings.Contains(err.Error(), "nope") {
		t.Fatalf("want error message to name the missing field, got %v", err)
	}
}

func TestRuntimeObject_GetReturnsSpecZero(t *testing.T) {
	bp := NewBlueprint("test.get_zero",
		Field{Name: "p", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
		Field{Name: "s", Spec: &SliceSpec{Type: "uint32"}},
		Field{Name: "opt", Spec: &PtrSpec{Type: "uint32"}},
	)
	ro := mustRuntimeObject(t, bp)

	// PrimitiveSpec: typed zero, not nil.
	got := ro.Get("p")
	if got == nil {
		t.Fatalf("Get on never-Set PrimitiveSpec field returned nil; want typed zero")
	}
	if _, ok := got.(*Uint32); !ok {
		t.Fatalf("Get on PrimitiveSpec uint32 field: want *Uint32, got %T", got)
	}

	// SliceSpec: empty RuntimeSlice, not nil.
	got = ro.Get("s")
	if got == nil {
		t.Fatalf("Get on never-Set SliceSpec field returned nil; want empty RuntimeSlice")
	}
	if _, ok := got.(*RuntimeSlice); !ok {
		t.Fatalf("Get on SliceSpec field: want *RuntimeSlice, got %T", got)
	}

	// PtrSpec: canonical &Nil{} absent marker, not nil interface.
	got = ro.Get("opt")
	if got == nil {
		t.Fatalf("Get on never-Set PtrSpec field returned nil interface; want *Nil")
	}
	if _, ok := got.(*Nil); !ok {
		t.Fatalf("Get on PtrSpec field: want *Nil, got %T", got)
	}
}

func TestNewRuntimeObject_Nil(t *testing.T) {
	ro, err := NewRuntimeObject(nil)
	if err != nil {
		t.Fatalf("NewRuntimeObject(nil) returned err: %v", err)
	}
	if ro == nil {
		t.Fatal("NewRuntimeObject(nil) returned nil RuntimeObject; want empty")
	}
	if got := ro.ObjectType(); got != "" {
		t.Fatalf("ObjectType on empty RO: want \"\", got %q", got)
	}
	// Get on any name must not panic; returns nil.
	if v := ro.Get("anything"); v != nil {
		t.Fatalf("Get on empty RO: want nil, got %v", v)
	}
	// Set on empty RO returns ErrFieldTypeMismatch (no fields exist) without panic.
	if err := ro.Set("anything", uint32(1)); !errors.Is(err, ErrFieldTypeMismatch) {
		t.Fatalf("Set on empty RO: want ErrFieldTypeMismatch, got %v", err)
	}
	// WriteTo emits zero bytes.
	var buf bytes.Buffer
	n, err := ro.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo on empty RO: %v", err)
	}
	if n != 0 || buf.Len() != 0 {
		t.Fatalf("WriteTo on empty RO: want 0 bytes, got n=%d len=%d", n, buf.Len())
	}
}

func TestRuntimeObject_SetRejectsMismatch(t *testing.T) {
	bp := NewBlueprint("test.reject",
		Field{Name: "n", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
	)
	ro := mustRuntimeObject(t, bp)

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
	ro := mustRuntimeObject(t, bp)

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

	src := mustRuntimeObject(t, bp)
	var buf bytes.Buffer
	_, err := src.WriteTo(&buf)
	if !errors.Is(err, ErrBlueprintNotFound) {
		t.Fatalf("encode: want ErrBlueprintNotFound, got %v", err)
	}

	dst := mustRuntimeObject(t, bp)
	_, err = dst.ReadFrom(&buf)
	if !errors.Is(err, ErrBlueprintNotFound) {
		t.Fatalf("decode: want ErrBlueprintNotFound, got %v", err)
	}
}

func TestRuntimeObject_UnregisteredMapValue_FailsSymmetrically(t *testing.T) {
	bp := NewBlueprint("test.map.missing",
		Field{Name: "m", Spec: &MapSpec{KeyType: "string16", ValueType: "definitely-not-registered"}},
	)

	src := mustRuntimeObject(t, bp)
	var buf bytes.Buffer
	_, err := src.WriteTo(&buf)
	if !errors.Is(err, ErrBlueprintNotFound) {
		t.Fatalf("encode: want ErrBlueprintNotFound, got %v", err)
	}

	dst := mustRuntimeObject(t, bp)
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

	a := mustRuntimeObject(t, bp)
	_ = a.Set("n", uint32(7))
	_ = a.Set("s", "x")

	b := mustRuntimeObject(t, bp)
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

func TestRuntimeObject_JSON_PrimitiveRoundTrip(t *testing.T) {
	bp := NewBlueprint("test.json.prim",
		Field{Name: "n", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
		Field{Name: "s", Spec: &PrimitiveSpec{PrimitiveType: "string16"}},
		Field{Name: "b", Spec: &PrimitiveSpec{PrimitiveType: "bool"}},
	)

	src := mustRuntimeObject(t, bp)
	_ = src.Set("n", uint32(42))
	_ = src.Set("s", "hello")
	_ = src.Set("b", true)

	data, err := json.Marshal(src)
	if err != nil {
		t.Fatal(err)
	}

	dst := mustRuntimeObject(t, bp)
	err = json.Unmarshal(data, dst)
	if err != nil {
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

func TestRuntimeObject_JSON_PtrSpec_NullVsValue(t *testing.T) {
	bp := NewBlueprint("test.json.ptr",
		Field{Name: "p", Spec: &PtrSpec{Type: "uint32"}},
	)

	// nil case: spec-zero is &Nil{}, marshals to null, round-trips back to *Nil
	srcNil := mustRuntimeObject(t, bp)
	dataNil, err := json.Marshal(srcNil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(dataNil), `"p":null`) {
		t.Fatalf("nil PtrSpec: want \"p\":null, got %s", dataNil)
	}
	dstNil := mustRuntimeObject(t, bp)
	if err := json.Unmarshal(dataNil, dstNil); err != nil {
		t.Fatal(err)
	}
	if _, ok := dstNil.Get("p").(*Nil); !ok {
		t.Fatalf("nil round-trip: want *Nil, got %#v", dstNil.Get("p"))
	}

	// non-nil case
	src := mustRuntimeObject(t, bp)
	_ = src.Set("p", NewUint32(7))
	data, err := json.Marshal(src)
	if err != nil {
		t.Fatal(err)
	}
	dst := mustRuntimeObject(t, bp)
	if err := json.Unmarshal(data, dst); err != nil {
		t.Fatal(err)
	}
	got, ok := dst.Get("p").(*Uint32)
	if !ok || *got != 7 {
		t.Fatalf("value round-trip: want 7, got %#v", dst.Get("p"))
	}
}

func TestRuntimeObject_JSON_SliceRoundTrip(t *testing.T) {
	bp := NewBlueprint("test.json.slice",
		Field{Name: "items", Spec: &SliceSpec{Type: "uint32"}},
	)

	src := mustRuntimeObject(t, bp)
	rs, err := NewRuntimeSlice("uint32")
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range []uint32{1, 2, 3} {
		_ = rs.Append(NewUint32(v))
	}
	_ = src.Set("items", rs)

	data, err := json.Marshal(src)
	if err != nil {
		t.Fatal(err)
	}

	dst := mustRuntimeObject(t, bp)
	if err := json.Unmarshal(data, dst); err != nil {
		t.Fatal(err)
	}

	got, ok := dst.Get("items").(*RuntimeSlice)
	if !ok || got.Len() != 3 {
		t.Fatalf("items: want len 3, got %#v", dst.Get("items"))
	}
	for i, want := range []uint32{1, 2, 3} {
		u, _ := got.At(i).(*Uint32)
		if u == nil || *u != Uint32(want) {
			t.Fatalf("[%d]: want %d, got %#v", i, want, got.At(i))
		}
	}
}

func TestRuntimeObject_JSON_MapRoundTrip(t *testing.T) {
	bp := NewBlueprint("test.json.map",
		Field{Name: "m", Spec: &MapSpec{KeyType: "string16", ValueType: "uint32"}},
	)

	src := mustRuntimeObject(t, bp)
	sm, err := NewRuntimeMap("string16", "uint32")
	if err != nil {
		t.Fatal(err)
	}
	_ = sm.Set("a", NewUint32(10))
	_ = sm.Set("b", NewUint32(20))
	_ = src.Set("m", sm)

	data, err := json.Marshal(src)
	if err != nil {
		t.Fatal(err)
	}

	dst := mustRuntimeObject(t, bp)
	if err := json.Unmarshal(data, dst); err != nil {
		t.Fatal(err)
	}

	got, ok := dst.Get("m").(*RuntimeMap)
	if !ok || got.Len() != 2 {
		t.Fatalf("m: want len 2, got %#v", dst.Get("m"))
	}
	for k, want := range map[string]uint32{"a": 10, "b": 20} {
		v, _ := got.Get(k)
		u, _ := v.(*Uint32)
		if u == nil || *u != Uint32(want) {
			t.Fatalf("m[%q]: want %d, got %#v", k, want, v)
		}
	}
}

func TestRuntimeObject_JSON_ObjectSpec_Polymorphic(t *testing.T) {
	bp := NewBlueprint("test.json.obj",
		Field{Name: "any", Spec: &ObjectSpec{}},
	)

	src := mustRuntimeObject(t, bp)
	_ = src.Set("any", NewString16("hi"))

	data, err := json.Marshal(src)
	if err != nil {
		t.Fatal(err)
	}
	// wire shape: {"any":{"Type":"string16","Object":"hi"}}
	if !strings.Contains(string(data), `"Type":"string16"`) {
		t.Fatalf("missing Type tag on ObjectSpec field: %s", data)
	}

	dst := mustRuntimeObject(t, bp)
	if err := json.Unmarshal(data, dst); err != nil {
		t.Fatal(err)
	}
	got, ok := dst.Get("any").(*String16)
	if !ok || *got != "hi" {
		t.Fatalf("any: want \"hi\" String16, got %#v", dst.Get("any"))
	}
}

func TestRuntimeObject_JSON_ObjectSpec_NullIsNil(t *testing.T) {
	bp := NewBlueprint("test.json.obj.null",
		Field{Name: "any", Spec: &ObjectSpec{}},
	)
	src := mustRuntimeObject(t, bp) // spec-zero leaves any=&Nil{}
	data, err := json.Marshal(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"any":null`) {
		t.Fatalf("want \"any\":null, got %s", data)
	}
	dst := mustRuntimeObject(t, bp)
	if err := json.Unmarshal(data, dst); err != nil {
		t.Fatal(err)
	}
	if _, ok := dst.Get("any").(*Nil); !ok {
		t.Fatalf("want *Nil, got %#v", dst.Get("any"))
	}
}

func TestRuntimeObject_JSON_MissingFieldsLeaveSpecZero(t *testing.T) {
	bp := NewBlueprint("test.json.missing",
		Field{Name: "kept", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
		Field{Name: "dropped", Spec: &PrimitiveSpec{PrimitiveType: "string16"}},
	)
	dst := mustRuntimeObject(t, bp)
	if err := json.Unmarshal([]byte(`{"kept":7}`), dst); err != nil {
		t.Fatal(err)
	}
	if u, _ := dst.Get("kept").(*Uint32); u == nil || *u != 7 {
		t.Fatalf("kept: want 7, got %#v", dst.Get("kept"))
	}
	if s, _ := dst.Get("dropped").(*String16); s == nil || *s != "" {
		t.Fatalf("dropped: want spec-zero String16, got %#v", dst.Get("dropped"))
	}
}

func TestRuntimeObject_JSON_NoBlueprintError(t *testing.T) {
	ro := &RuntimeObject{} // bp == nil
	err := json.Unmarshal([]byte(`{"x":1}`), ro)
	if err == nil || !strings.Contains(err.Error(), "no Blueprint") {
		t.Fatalf("want no-Blueprint error, got %v", err)
	}
}

// TestRuntimeObject_DecodeDepthCap_MutualPtr proves the decode depth cap converts
// a mutually-recursive Blueprint pair (X.PtrSpec→Y, Y.PtrSpec→X) into a typed
// ErrDepthExceeded instead of a stack overflow when fed a stream that keeps every
// presence byte non-zero.
//
// why direct-Set bypass: RegisterBlueprint now enforces dependency closure, which
// forbids PtrSpec cycles through the public API. The decode-side safety net is still
// required for cycles formed via ObjectSpec or heterogeneous containers, and for
// pre-existing registry states; install the pair directly to exercise it.
func TestRuntimeObject_DecodeDepthCap_MutualPtr(t *testing.T) {
	xBP := NewBlueprint("test.depth.x",
		Field{Name: "y", Spec: &PtrSpec{Type: "test.depth.y"}},
	)
	yBP := NewBlueprint("test.depth.y",
		Field{Name: "x", Spec: &PtrSpec{Type: "test.depth.x"}},
	)
	DefaultBlueprints().Blueprints.Set("test.depth.x", xBP)
	DefaultBlueprints().Blueprints.Set("test.depth.y", yBP)

	ro := mustRuntimeObject(t, xBP)
	// All-ones stream: every PtrSpec presence byte decodes as "present", forcing the
	// decoder to recurse into the referenced type indefinitely (until the depth cap).
	r := bytes.NewReader(bytes.Repeat([]byte{0xFF}, 4*MaxBlueprintDepth))
	_, err := ro.ReadFrom(r)
	if !errors.Is(err, ErrDepthExceeded) {
		t.Fatalf("want ErrDepthExceeded, got %v", err)
	}
}

// TestRuntimeObject_DepthCap_Boundary — §3.4. Decode at depth MaxBlueprintDepth must
// succeed; decode at depth MaxBlueprintDepth+1 must surface ErrDepthExceeded.
//
// Frame accounting: each RuntimeObject.ReadFrom increments depth on entry. A chain of
// N nested PtrSpec frames needs (N-1) 0x01 presence bytes followed by one 0x00 (final
// frame has no further recursion). For overflow, frame N is reached and ErrDepthExceeded
// fires when the (N+1)-th frame's enter() runs.
func TestRuntimeObject_DepthCap_Boundary(t *testing.T) {
	xBP := NewBlueprint("test.depth.boundary.x",
		Field{Name: "y", Spec: &PtrSpec{Type: "test.depth.boundary.y"}},
	)
	yBP := NewBlueprint("test.depth.boundary.y",
		Field{Name: "x", Spec: &PtrSpec{Type: "test.depth.boundary.x"}},
	)
	// why: same closure-bypass rationale as TestRuntimeObject_DecodeDepthCap_MutualPtr.
	DefaultBlueprints().Blueprints.Set("test.depth.boundary.x", xBP)
	DefaultBlueprints().Blueprints.Set("test.depth.boundary.y", yBP)

	t.Run("at_cap_succeeds", func(t *testing.T) {
		// MaxBlueprintDepth frames: (MaxBlueprintDepth-1) present bytes, then one absent.
		payload := append(
			bytes.Repeat([]byte{0x01}, MaxBlueprintDepth-1),
			0x00,
		)
		ro := mustRuntimeObject(t, xBP)
		if _, err := ro.ReadFrom(bytes.NewReader(payload)); err != nil {
			t.Fatalf("decode at depth %d should succeed, got %v", MaxBlueprintDepth, err)
		}
	})

	t.Run("over_cap_errors", func(t *testing.T) {
		// All present: each frame recurses, so after MaxBlueprintDepth frames the next
		// enter() exceeds the cap.
		payload := bytes.Repeat([]byte{0x01}, MaxBlueprintDepth+1)
		ro := mustRuntimeObject(t, xBP)
		_, err := ro.ReadFrom(bytes.NewReader(payload))
		if !errors.Is(err, ErrDepthExceeded) {
			t.Fatalf("decode past cap should return ErrDepthExceeded, got %v", err)
		}
	})
}

// §3.5 — RefSpec zero-byte recursion at decode-time — intentionally NOT written.
//
// During investigation we found that constructing a RuntimeObject for a Blueprint with
// mutually recursive RefSpec fields (A.b=RefSpec→B, B.a=RefSpec→A) crashes the test
// process via stack overflow inside NewRuntimeObject → specZero → New → NewRuntimeObject
// (runtime_object.go:48,157), *before* any depth-checked I/O runs. The MaxBlueprintDepth
// cap only wraps ReadFrom/WriteTo (depth.go:25,35), so the recursion is unguarded at
// construction. Writing this test requires either (1) construction-time cycle detection
// in validateBlueprint (today it rejects only self-reference, not two-step cycles), or
// (2) lazy spec-zero. Both are production-code changes deferred from this iteration.
