package astral

import (
	"bytes"
	"slices"
	"testing"
)

func TestSliceOfInts(t *testing.T) {
	var src, dst []int

	for i := 0; i < 10; i++ {
		src = append(src, i)
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
