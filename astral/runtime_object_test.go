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

// TestRuntimeObject_UnregisteredArrayElement_BothEndsError pins that an unregistered element
// type fails both encode and decode, with the failure classes now divergent:
//   - specZero swallows the constructor error and stores nil (runtime_object.go:218-225),
//     so encode hits writeField with value==nil and surfaces "nil value for ArraySpec".
//   - readField still resolves through newRuntimeArrayWith and surfaces ErrBlueprintNotFound.
//
// The symmetric-error guarantee (writeField re-resolving against defaultBlueprints) was
// dropped because it broke per-call-registry encode without catching anything decode wouldn't.
func TestRuntimeObject_UnregisteredArrayElement_BothEndsError(t *testing.T) {
	bp := NewBlueprint("test.array.missing",
		Field{Name: "a", Spec: &ArraySpec{Type: "definitely-not-registered", Length: 2}},
	)

	src := mustRuntimeObject(t, bp)
	var buf bytes.Buffer
	if _, err := src.WriteTo(&buf); err == nil {
		t.Fatal("encode: expected error, got nil")
	}

	dst := mustRuntimeObject(t, bp)
	_, err := dst.ReadFrom(&buf)
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

// TestRuntimeObject_UnregisteredSliceElement_BothEndsError pins the post-defensive-drop
// contract: specZero stores nil for an unregistered element type, so encode fails with
// "nil value for SliceSpec"; decode still surfaces ErrBlueprintNotFound via newRuntimeSliceWith.
// The previous symmetric ErrBlueprintNotFound on encode was driven by a defaultBlueprints-bound
// re-resolve in writeField that broke per-call-registry callers.
func TestRuntimeObject_UnregisteredSliceElement_BothEndsError(t *testing.T) {
	bp := NewBlueprint("test.slice.missing",
		Field{Name: "items", Spec: &SliceSpec{Type: "definitely-not-registered"}},
	)

	src := mustRuntimeObject(t, bp)
	var buf bytes.Buffer
	if _, err := src.WriteTo(&buf); err == nil {
		t.Fatal("encode: expected error, got nil")
	}

	dst := mustRuntimeObject(t, bp)
	_, err := dst.ReadFrom(&buf)
	if !errors.Is(err, ErrBlueprintNotFound) {
		t.Fatalf("decode: want ErrBlueprintNotFound, got %v", err)
	}
}

func TestRuntimeObject_UnregisteredMapValue_BothEndsError(t *testing.T) {
	bp := NewBlueprint("test.map.missing",
		Field{Name: "m", Spec: &MapSpec{KeyType: "string16", ValueType: "definitely-not-registered"}},
	)

	src := mustRuntimeObject(t, bp)
	var buf bytes.Buffer
	if _, err := src.WriteTo(&buf); err == nil {
		t.Fatal("encode: expected error, got nil")
	}

	dst := mustRuntimeObject(t, bp)
	_, err := dst.ReadFrom(&buf)
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

// TestRuntimeObject_JSON_CaseInsensitive pins the Postel posture chosen for the JSON
// decoder: incoming keys match field names case-insensitively (mirroring encoding/json
// and structValue), excess keys are silently ignored, and two payload keys folding to
// the same lowercase form are rejected as ambiguous.
func TestRuntimeObject_JSON_CaseInsensitive(t *testing.T) {
	bp := NewBlueprint("test.json.case",
		Field{Name: "SomeNumber", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
		Field{Name: "UserName", Spec: &PrimitiveSpec{PrimitiveType: "string16"}},
	)

	cases := []struct {
		name     string
		payload  string
		wantNum  uint32
		wantName String16
		wantErr  string // substring; "" means expect success
	}{
		{"exact", `{"SomeNumber":42,"UserName":"alice"}`, 42, "alice", ""},
		{"all-lower", `{"somenumber":42,"username":"alice"}`, 42, "alice", ""},
		{"all-upper", `{"SOMENUMBER":42,"USERNAME":"alice"}`, 42, "alice", ""},
		{"mixed", `{"sOmEnUmBeR":42,"uSeRnAmE":"alice"}`, 42, "alice", ""},
		{"excess-ignored", `{"SomeNumber":42,"UserName":"alice","extra":"x"}`, 42, "alice", ""},
		{"missing-leaves-zero", `{"SomeNumber":42}`, 42, "", ""},
		{"case-collision-rejected", `{"someNumber":1,"SOMENUMBER":2}`, 0, "", "duplicate fields due to case insensitivity"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dst := mustRuntimeObject(t, bp)
			err := json.Unmarshal([]byte(c.payload), dst)
			if c.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), c.wantErr) {
					t.Fatalf("want err containing %q, got %v", c.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if u, _ := dst.Get("SomeNumber").(*Uint32); u == nil || *u != Uint32(c.wantNum) {
				t.Fatalf("SomeNumber: want %d, got %#v", c.wantNum, dst.Get("SomeNumber"))
			}
			if s, _ := dst.Get("UserName").(*String16); s == nil || *s != c.wantName {
				t.Fatalf("UserName: want %q, got %#v", c.wantName, dst.Get("UserName"))
			}
		})
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
	DefaultBlueprints().entries.Set("test.depth.x", xBP)
	DefaultBlueprints().entries.Set("test.depth.y", yBP)

	ro := mustRuntimeObject(t, xBP)
	// All-0x01 stream: every PtrSpec presence byte decodes as the canonical "present"
	// flag (readField now enforces strict 0/1), forcing the decoder to recurse into the
	// referenced type indefinitely (until the depth cap).
	r := bytes.NewReader(bytes.Repeat([]byte{0x01}, 4*MaxBlueprintDepth))
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
	DefaultBlueprints().entries.Set("test.depth.boundary.x", xBP)
	DefaultBlueprints().entries.Set("test.depth.boundary.y", yBP)

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

// TestPtrSpec_WriteNilForms_AllAbsent pins the contract aligned with ptrValue.IsNil():
// three "absent" forms — interface-nil, *Nil, and typed nil pointer — all serialize as
// presence=0 (single 0x00 byte). Before the isPtrNil unification a typed nil pointer
// bypassed both checks, wrote presence=1, then crashed in value.WriteTo.
func TestPtrSpec_WriteNilForms_AllAbsent(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	if _, err := bps.RegisterBlueprint(NewBlueprintAlias("test.ptr_payload", "uint8")); err != nil {
		t.Fatal(err)
	}
	bp := NewBlueprint("test.ptr_envelope",
		Field{Name: "p", Spec: &PtrSpec{Type: "test.ptr_payload"}},
	)
	if _, err := bps.RegisterBlueprint(bp); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name  string
		value Object
	}{
		{"interface-nil", nil},
		{"Nil marker", &Nil{}},
		{"typed nil ptr", (*RuntimeObject)(nil)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ro := bps.New("test.ptr_envelope").(*RuntimeObject)
			ro.fields[0].Value = c.value
			var buf bytes.Buffer
			if _, err := ro.WriteTo(&buf); err != nil {
				t.Fatalf("WriteTo: %v", err)
			}
			if buf.Len() != 1 || buf.Bytes()[0] != 0 {
				t.Fatalf("want presence=0 (single 0x00 byte), got %v", buf.Bytes())
			}
		})
	}
}

// TestPtrSpec_MarshalJSON_NilFormsAllNull pins that PtrSpec JSON emits null for the same
// three absent forms — matches the binary nil-flag protocol so both encodings stay
// symmetric.
func TestPtrSpec_MarshalJSON_NilFormsAllNull(t *testing.T) {
	bps := NewBlueprints(DefaultBlueprints())
	if _, err := bps.RegisterBlueprint(NewBlueprintAlias("test.ptr_payload_json", "uint8")); err != nil {
		t.Fatal(err)
	}
	bp := NewBlueprint("test.ptr_envelope_json",
		Field{Name: "p", Spec: &PtrSpec{Type: "test.ptr_payload_json"}},
	)
	if _, err := bps.RegisterBlueprint(bp); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name  string
		value Object
	}{
		{"interface-nil", nil},
		{"Nil marker", &Nil{}},
		{"typed nil ptr", (*RuntimeObject)(nil)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ro := bps.New("test.ptr_envelope_json").(*RuntimeObject)
			ro.fields[0].Value = c.value
			data, err := json.Marshal(ro)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			if !bytes.Contains(data, []byte(`"p":null`)) {
				t.Fatalf("want field p as null, got %s", data)
			}
		})
	}
}

// TestObjectSpec_NestedRegistry_HonorsWithBlueprints pins the ObjectSpec branch of
// readField: the polymorphic envelope recursively calls Decode(dr, WithBlueprints(bps)),
// so the inner Decode must resolve the embedded type tag via the caller's per-call
// registry (threaded through objectReader.bps), not the package-level default. Mirrors
// TestRefSpec_ToAlias_EndToEnd for the ObjectSpec field shape.
func TestObjectSpec_NestedRegistry_HonorsWithBlueprints(t *testing.T) {
	// local: typed Go value + a Blueprint with a polymorphic ObjectSpec field
	local := NewBlueprints(DefaultBlueprints())
	if err := local.Add(newPrimitiveAliasMode()); err != nil {
		t.Fatal(err)
	}
	if _, err := local.RegisterBlueprint(NewBlueprint("test.obj_env",
		Field{Name: "m", Spec: &ObjectSpec{}},
	)); err != nil {
		t.Fatal(err)
	}
	src := local.New("test.obj_env").(*RuntimeObject)
	mode := aliasableMode(3)
	if err := src.Set("m", &mode); err != nil {
		t.Fatal(err)
	}
	buf, err := EncodeBytes(src, WithBlueprints(local))
	if err != nil {
		t.Fatal(err)
	}

	// remote: alias + same envelope Blueprint, no Go type for Mode
	remote := NewBlueprints(DefaultBlueprints())
	if _, err := remote.RegisterBlueprint(NewBlueprintAlias("test.aliasable_mode", "uint8")); err != nil {
		t.Fatal(err)
	}
	if _, err := remote.RegisterBlueprint(NewBlueprint("test.obj_env",
		Field{Name: "m", Spec: &ObjectSpec{}},
	)); err != nil {
		t.Fatal(err)
	}
	got, _, err := Decode(bytes.NewReader(buf), WithBlueprints(remote))
	if err != nil {
		t.Fatal(err)
	}
	ro, ok := got.(*RuntimeObject)
	if !ok {
		t.Fatalf("want *RuntimeObject, got %T", got)
	}
	field := ro.Get("m")
	ra, ok := field.(*RuntimeObject)
	if !ok {
		t.Fatalf("field m: want *RuntimeObject, got %T", field)
	}
	u, ok := ra.Underlying().(*Uint8)
	if !ok || uint8(*u) != 3 {
		t.Fatalf("field m underlying drift: %#v", ra.Underlying())
	}
}

// TestBlueprintsNew_SpecZero_HonorsCustomBlueprints pins the construction-side counterpart
// to TestObjectSpec_NestedRegistry_HonorsWithBlueprints. Blueprints.New must seed each field's
// spec-zero through the same registry it dispatched through, so a RefSpec/SliceSpec/MapSpec
// pointing at a type registered only in a child Blueprints materializes as the typed zero
// instead of interface-nil (which would later surface as a misleading "nil value for spec"
// error in writeField). Without threading bps through specZero, the package-level default
// is consulted and the lookup silently fails.
func TestBlueprintsNew_SpecZero_HonorsCustomBlueprints(t *testing.T) {
	local := NewBlueprints(DefaultBlueprints())
	if _, err := local.RegisterBlueprint(NewBlueprintAlias("test.local_mode", "uint8")); err != nil {
		t.Fatal(err)
	}
	if _, err := local.RegisterBlueprint(NewBlueprint("test.local_env",
		Field{Name: "m", Spec: &RefSpec{Type: "test.local_mode"}},
		Field{Name: "s", Spec: &SliceSpec{Type: "test.local_mode"}},
	)); err != nil {
		t.Fatal(err)
	}

	ro, ok := local.New("test.local_env").(*RuntimeObject)
	if !ok {
		t.Fatalf("local.New: want *RuntimeObject, got %T", local.New("test.local_env"))
	}

	// RefSpec spec-zero must be a *RuntimeObject from local, not interface-nil.
	got := ro.Get("m")
	if got == nil {
		t.Fatal("Get on RefSpec field returned nil; want typed zero from custom registry")
	}
	if _, ok := got.(*RuntimeObject); !ok {
		t.Fatalf("RefSpec spec-zero: want *RuntimeObject, got %T", got)
	}

	// SliceSpec spec-zero must be an empty *RuntimeSlice whose elemName resolved through local.
	got = ro.Get("s")
	if got == nil {
		t.Fatal("Get on SliceSpec field returned nil; want empty RuntimeSlice from custom registry")
	}
	if _, ok := got.(*RuntimeSlice); !ok {
		t.Fatalf("SliceSpec spec-zero: want *RuntimeSlice, got %T", got)
	}
}

// TestRuntimeObject_Encode_HonorsCustomBlueprints_Slice pins the encode-side counterpart of
// TestBlueprintsNew_SpecZero_HonorsCustomBlueprints: a RuntimeObject built from a child
// registry, whose SliceSpec element type lives only in that registry, must encode without
// hitting a defaultBlueprints-bound re-resolution. Before dropping writeField's defensive
// re-resolves (and exposing NewRuntimeSliceWith), encode failed with ErrBlueprintNotFound
// even though decode worked fine.
func TestRuntimeObject_Encode_HonorsCustomBlueprints_Slice(t *testing.T) {
	local := NewBlueprints(DefaultBlueprints())
	if _, err := local.RegisterBlueprint(NewBlueprint("test.encode_payload",
		Field{Name: "n", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
	)); err != nil {
		t.Fatal(err)
	}
	if _, err := local.RegisterBlueprint(NewBlueprint("test.encode_envelope",
		Field{Name: "items", Spec: &SliceSpec{Type: "test.encode_payload"}},
	)); err != nil {
		t.Fatal(err)
	}

	src := local.New("test.encode_envelope").(*RuntimeObject)
	items, err := NewRuntimeSliceWith(local, "test.encode_payload")
	if err != nil {
		t.Fatal(err)
	}
	payload := local.New("test.encode_payload").(*RuntimeObject)
	if err := payload.Set("n", uint32(42)); err != nil {
		t.Fatal(err)
	}
	if err := items.Append(payload); err != nil {
		t.Fatal(err)
	}
	if err := src.Set("items", items); err != nil {
		t.Fatal(err)
	}

	buf, err := EncodeBytes(src, WithBlueprints(local))
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	got, _, err := Decode(bytes.NewReader(buf), WithBlueprints(local))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	dst, ok := got.(*RuntimeObject)
	if !ok {
		t.Fatalf("want *RuntimeObject, got %T", got)
	}
	rs, ok := dst.Get("items").(*RuntimeSlice)
	if !ok || rs.Len() != 1 {
		t.Fatalf("decoded items: want one element, got %#v", dst.Get("items"))
	}
	elem, ok := rs.At(0).(*RuntimeObject)
	if !ok {
		t.Fatalf("decoded element: want *RuntimeObject, got %T", rs.At(0))
	}
	if n, _ := elem.Get("n").(*Uint32); n == nil || *n != 42 {
		t.Fatalf("decoded element.n: want 42, got %#v", elem.Get("n"))
	}
}
