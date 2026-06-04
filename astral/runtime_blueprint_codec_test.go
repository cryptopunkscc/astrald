package astral

import (
	"bytes"
	"testing"
)

// TestRuntimeSlice_RuntimeBlueprintElement_RoundTrip pins the slow-path codec for a
// SliceSpec whose element type is a runtime Blueprint. Pre-fix the generic codec
// allocated unbound *RuntimeObject via reflect.New, so per-element ReadFrom returned
// (0, nil) — the receiver consumed only the length prefix and per-element nil-flags
// while leaving every element payload on the wire.
func TestRuntimeSlice_RuntimeBlueprintElement_RoundTrip(t *testing.T) {
	userBP := NewBlueprint("test.codec.slice.user",
		Field{Name: "n", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
	)
	_, err := RegisterBlueprint(userBP)
	if err != nil {
		t.Fatal(err)
	}

	src, err := NewRuntimeSlice("test.codec.slice.user")
	if err != nil {
		t.Fatal(err)
	}
	values := []uint32{42, 7, 99}
	for _, v := range values {
		u := New("test.codec.slice.user").(*RuntimeObject)
		err := u.Set("n", v)
		if err != nil {
			t.Fatal(err)
		}
		err = src.Append(u)
		if err != nil {
			t.Fatal(err)
		}
	}

	var buf bytes.Buffer
	_, err = src.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}

	dst, err := NewRuntimeSlice("test.codec.slice.user")
	if err != nil {
		t.Fatal(err)
	}
	// why: assert the decoder consumed the entire wire (buf.Len() after ReadFrom == 0).
	// Pre-fix, decode silently no-op'd per element and left the bulk of the payload on
	// the wire — that's the regression this test pins. Byte-count parity (read == wrote)
	// is left off because primitive readers under-report n today.
	_, err = dst.ReadFrom(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 0 {
		t.Fatalf("decoder left %d bytes on the wire", buf.Len())
	}
	if dst.Len() != len(values) {
		t.Fatalf("len: want %d, got %d", len(values), dst.Len())
	}
	for i, want := range values {
		ro, ok := dst.At(i).(*RuntimeObject)
		if !ok {
			t.Fatalf("elem %d: want *RuntimeObject, got %T", i, dst.At(i))
		}
		got, ok := ro.Get("n").(*Uint32)
		if !ok || uint32(*got) != want {
			t.Fatalf("elem %d.n: want %d, got %#v", i, want, ro.Get("n"))
		}
	}
}

// TestRuntimeArray_RuntimeBlueprintElement_RoundTrip is the same regression for ArraySpec.
// The arrayValue codec has no length prefix (length is in the schema), so the pre-fix
// failure is per-element: each fresh *RuntimeObject{bp:nil} consumed only the nil-flag.
func TestRuntimeArray_RuntimeBlueprintElement_RoundTrip(t *testing.T) {
	itemBP := NewBlueprint("test.codec.array.item",
		Field{Name: "v", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
	)
	_, err := RegisterBlueprint(itemBP)
	if err != nil {
		t.Fatal(err)
	}

	const N = 3
	src, err := NewRuntimeArray("test.codec.array.item", N)
	if err != nil {
		t.Fatal(err)
	}
	values := [N]uint32{10, 20, 30}
	for i, v := range values {
		item := New("test.codec.array.item").(*RuntimeObject)
		err := item.Set("v", v)
		if err != nil {
			t.Fatal(err)
		}
		err = src.Set(i, item)
		if err != nil {
			t.Fatal(err)
		}
	}

	var buf bytes.Buffer
	_, err = src.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}

	dst, err := NewRuntimeArray("test.codec.array.item", N)
	if err != nil {
		t.Fatal(err)
	}
	_, err = dst.ReadFrom(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 0 {
		t.Fatalf("decoder left %d bytes on the wire", buf.Len())
	}
	for i, want := range values {
		ro, ok := dst.At(i).(*RuntimeObject)
		if !ok {
			t.Fatalf("elem %d: want *RuntimeObject, got %T", i, dst.At(i))
		}
		got, ok := ro.Get("v").(*Uint32)
		if !ok || uint32(*got) != want {
			t.Fatalf("elem %d.v: want %d, got %#v", i, want, ro.Get("v"))
		}
	}
}

// TestRuntimeMap_RuntimeBlueprintValue_RoundTrip is the same regression for MapSpec.
// Keys are primitive (string16) and unaffected; the value side is what was silently
// no-op'd. Iteration order is map-dependent, so verify by lookup.
func TestRuntimeMap_RuntimeBlueprintValue_RoundTrip(t *testing.T) {
	entryBP := NewBlueprint("test.codec.map.entry",
		Field{Name: "x", Spec: &PrimitiveSpec{PrimitiveType: "uint32"}},
	)
	_, err := RegisterBlueprint(entryBP)
	if err != nil {
		t.Fatal(err)
	}

	src, err := NewRuntimeMap("string16", "test.codec.map.entry")
	if err != nil {
		t.Fatal(err)
	}
	entries := map[string]uint32{"a": 1, "bb": 2, "ccc": 3}
	for k, v := range entries {
		ro := New("test.codec.map.entry").(*RuntimeObject)
		err := ro.Set("x", v)
		if err != nil {
			t.Fatal(err)
		}
		err = src.Set(k, ro)
		if err != nil {
			t.Fatal(err)
		}
	}

	var buf bytes.Buffer
	_, err = src.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}

	dst, err := NewRuntimeMap("string16", "test.codec.map.entry")
	if err != nil {
		t.Fatal(err)
	}
	_, err = dst.ReadFrom(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 0 {
		t.Fatalf("decoder left %d bytes on the wire", buf.Len())
	}
	if dst.Len() != len(entries) {
		t.Fatalf("len: want %d, got %d", len(entries), dst.Len())
	}
	for k, want := range entries {
		v, ok := dst.Get(k)
		if !ok {
			t.Fatalf("missing key %q", k)
		}
		ro, ok := v.(*RuntimeObject)
		if !ok {
			t.Fatalf("key %q: want *RuntimeObject, got %T", k, v)
		}
		got, ok := ro.Get("x").(*Uint32)
		if !ok || uint32(*got) != want {
			t.Fatalf("key %q.x: want %d, got %#v", k, want, ro.Get("x"))
		}
	}
}
