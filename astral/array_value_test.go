package astral

import (
	"bytes"
	"slices"
	"testing"
)

func TestArrayOfInts(t *testing.T) {
	var src, dst [10]int

	for i := 0; i < 10; i++ {
		src[i] = i
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

	if slices.Compare(dst[:], src[:]) != 0 {
		t.Fatal("slice values not equal")
	}
	dst = [10]int{}

	// test json marshaling
	json, err := srcObject.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	err = dstObject.UnmarshalJSON(json)
	if err != nil {
		t.Fatal(err)
	}

	if slices.Compare(dst[:], src[:]) != 0 {
		t.Fatal("slice values not equal")
	}
}

func TestArray_AllZero_RoundTrip(t *testing.T) {
	var src, dst [3]uint32

	srcObject := Objectify(&src)
	var buf bytes.Buffer
	n, err := srcObject.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	// Length is part of the schema → no count prefix; 3 uint32 = 12 bytes.
	if n != 12 || buf.Len() != 12 {
		t.Fatalf("want 12 bytes for [3]uint32, got n=%d len=%d", n, buf.Len())
	}

	dstObject := Objectify(&dst)
	if _, err := dstObject.ReadFrom(&buf); err != nil {
		t.Fatal(err)
	}
	if slices.Compare(dst[:], src[:]) != 0 {
		t.Fatalf("round-trip mismatch: %v vs %v", dst, src)
	}
}

func TestRuntimeArray_SetOutOfBounds(t *testing.T) {
	ra, err := NewRuntimeArray("uint32", 3)
	if err != nil {
		t.Fatal(err)
	}
	v := Uint32(7)
	if err := ra.Set(5, &v); err == nil {
		t.Fatal("want error for Set(5) on length-3 array, got nil")
	}
	if err := ra.Set(-1, &v); err == nil {
		t.Fatal("want error for Set(-1), got nil")
	}
	// Set at last valid index must succeed.
	if err := ra.Set(2, &v); err != nil {
		t.Fatalf("Set at last valid index: %v", err)
	}
}
