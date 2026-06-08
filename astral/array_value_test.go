package astral

import (
	"bytes"
	"encoding/hex"
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
	// Length is part of the schema → no count prefix. 3 elements × (1-byte presence + 4-byte
	// uint32) = 15 bytes. The presence byte keeps value-typed [N]T wire identical to [N]*T.
	if n != 15 || buf.Len() != 15 {
		t.Fatalf("want 15 bytes for [3]uint32, got n=%d len=%d", n, buf.Len())
	}

	dstObject := Objectify(&dst)
	if _, err := dstObject.ReadFrom(&buf); err != nil {
		t.Fatal(err)
	}
	if slices.Compare(dst[:], src[:]) != 0 {
		t.Fatalf("round-trip mismatch: %v vs %v", dst, src)
	}
}

// TestArray_CrossCodecParity_ValueVsPtr — array equivalent of the slice value-vs-ptr parity
// guard. [2]Uint32, [2]*Uint32, and RuntimeArray("uint32", 2) must produce identical wire.
// No count prefix (length is in the schema).
func TestArray_CrossCodecParity_ValueVsPtr(t *testing.T) {
	wantHex := "01" + "00000001" + "01" + "00000002"

	valueForm := [2]Uint32{1, 2}
	var valBuf bytes.Buffer
	if _, err := Objectify(&valueForm).WriteTo(&valBuf); err != nil {
		t.Fatal(err)
	}

	ptrForm := [2]*Uint32{NewUint32(1), NewUint32(2)}
	var ptrBuf bytes.Buffer
	if _, err := Objectify(&ptrForm).WriteTo(&ptrBuf); err != nil {
		t.Fatal(err)
	}

	ra, err := NewRuntimeArray("uint32", 2)
	if err != nil {
		t.Fatal(err)
	}
	for i, v := range ptrForm {
		if err := ra.Set(i, v); err != nil {
			t.Fatal(err)
		}
	}
	var raBuf bytes.Buffer
	if _, err := ra.WriteTo(&raBuf); err != nil {
		t.Fatal(err)
	}

	gotVal := hex.EncodeToString(valBuf.Bytes())
	gotPtr := hex.EncodeToString(ptrBuf.Bytes())
	gotRA := hex.EncodeToString(raBuf.Bytes())
	if gotVal != wantHex || gotPtr != wantHex || gotRA != wantHex {
		t.Fatalf("wire mismatch:\n want: %s\n [N]T: %s\n[N]*T: %s\n   RA: %s",
			wantHex, gotVal, gotPtr, gotRA)
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
