package astral

import (
	"bytes"
	"slices"
	"testing"
)

func TestSliceOfInts(t *testing.T) {
	var src, dst []int64

	for i := 0; i < 10; i++ {
		src = append(src, int64(i))
	}

	srcObject := Objectify(&src)

	// test binary marshaling
	var buf = &bytes.Buffer{}

	_, err := srcObject.WriteTo(buf)
	if err != nil {
		t.Fatal(err)
	}

	dstObject := Objectify(&dst)
	_, err = dstObject.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}

	if slices.Compare(dst, src) != 0 {
		t.Fatal("slice values not equal")
	}
	dst = nil

	// test json marshaling
	json, err := srcObject.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	err = dstObject.UnmarshalJSON(json)
	if err != nil {
		t.Fatal(err)
	}

	if slices.Compare(dst, src) != 0 {
		t.Fatal("slice values not equal")
	}
}

// TestSlice_Empty_RoundTrip — §4.1. Empty slices exercise the count=0 short-circuit
// across element types.
func TestSlice_Empty_RoundTrip(t *testing.T) {
	t.Run("uint32", func(t *testing.T) {
		src := []uint32{}
		dst := []uint32{1, 2, 3} // pre-populated to verify it's cleared
		var buf bytes.Buffer
		if _, err := Objectify(&src).WriteTo(&buf); err != nil {
			t.Fatal(err)
		}
		// 4-byte count of zero, no payload.
		if buf.Len() != 4 {
			t.Fatalf("empty slice should encode to 4 bytes, got %d", buf.Len())
		}
		if _, err := Objectify(&dst).ReadFrom(&buf); err != nil {
			t.Fatal(err)
		}
		if len(dst) != 0 {
			t.Fatalf("decoded slice should be empty, got len=%d", len(dst))
		}
	})
	t.Run("string", func(t *testing.T) {
		src := []string{}
		dst := []string{"x"}
		var buf bytes.Buffer
		if _, err := Objectify(&src).WriteTo(&buf); err != nil {
			t.Fatal(err)
		}
		if _, err := Objectify(&dst).ReadFrom(&buf); err != nil {
			t.Fatal(err)
		}
		if len(dst) != 0 {
			t.Fatalf("decoded slice should be empty, got len=%d", len(dst))
		}
	})
	t.Run("object_interface", func(t *testing.T) {
		src := []Object{}
		dst := []Object{NewString16("preexisting")}
		var buf bytes.Buffer
		if _, err := Objectify(&src).WriteTo(&buf); err != nil {
			t.Fatal(err)
		}
		if _, err := Objectify(&dst).ReadFrom(&buf); err != nil {
			t.Fatal(err)
		}
		if len(dst) != 0 {
			t.Fatalf("decoded slice should be empty, got len=%d", len(dst))
		}
	})
}

// TestSlice_Heterogeneous_RoundTrip — §4.3. Slice element type Object (interface{})
// requires per-element type tag; mixed concrete types must round-trip.
func TestSlice_Heterogeneous_RoundTrip(t *testing.T) {
	src := []Object{
		NewString16("hello"),
		func() Object { v := Uint32(42); return &v }(),
		func() Object { v := Bool(true); return &v }(),
	}
	var dst []Object

	var buf bytes.Buffer
	if _, err := Objectify(&src).WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	if _, err := Objectify(&dst).ReadFrom(&buf); err != nil {
		t.Fatal(err)
	}

	if len(dst) != len(src) {
		t.Fatalf("len: want %d, got %d", len(src), len(dst))
	}
	for i, want := range src {
		if dst[i].ObjectType() != want.ObjectType() {
			t.Fatalf("element %d type: want %s, got %s", i, want.ObjectType(), dst[i].ObjectType())
		}
	}
	if s, ok := dst[0].(*String16); !ok || *s != "hello" {
		t.Fatalf("element 0: want String16 hello, got %#v", dst[0])
	}
	if u, ok := dst[1].(*Uint32); !ok || *u != 42 {
		t.Fatalf("element 1: want Uint32 42, got %#v", dst[1])
	}
	if b, ok := dst[2].(*Bool); !ok || !bool(*b) {
		t.Fatalf("element 2: want Bool true, got %#v", dst[2])
	}
}

// TestSlice_UnmarshalJSON_ShrinksSlice — §4.5. The decoder rebuilds the slice from the
// JSON array length, so a pre-existing larger slice must have its tail dropped — no stale
// tail elements.
func TestSlice_UnmarshalJSON_ShrinksSlice(t *testing.T) {
	dst := make([]uint32, 10)
	for i := range dst {
		dst[i] = 999
	}
	o := Objectify(&dst)
	if err := o.UnmarshalJSON([]byte(`[1,2,3]`)); err != nil {
		t.Fatal(err)
	}
	if len(dst) != 3 {
		t.Fatalf("len: want 3, got %d", len(dst))
	}
	for i, want := range []uint32{1, 2, 3} {
		if dst[i] != want {
			t.Fatalf("dst[%d]: want %d, got %d", i, want, dst[i])
		}
	}
}
