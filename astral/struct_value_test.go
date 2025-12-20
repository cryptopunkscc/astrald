package astral

import (
	"bytes"
	"slices"
	"testing"
)

type testStruct struct {
	SomeNumber int
	SomeList   []string
}

func TestStruct(t *testing.T) {
	var err error
	var src, dst testStruct

	src.SomeNumber = 123
	src.SomeList = []string{"hello", "world"}

	srcObject := Objectify(&src)

	var buf = &bytes.Buffer{}
	_, err = srcObject.WriteTo(buf)
	if err != nil {
		t.Fatal(err)
	}

	dstObject := Objectify(&dst)
	_, err = dstObject.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}

	switch {
	case src.SomeNumber != dst.SomeNumber:
		t.Fatal("numbers not equal")
	case slices.Compare(src.SomeList, dst.SomeList) != 0:
		t.Fatal("lists not equal")
	}

	dst = testStruct{}
	jdata, err := srcObject.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	dstObject = Objectify(&dst)
	err = dstObject.UnmarshalJSON(jdata)
	if err != nil {
		t.Fatal(err)
	}

	switch {
	case src.SomeNumber != dst.SomeNumber:
		t.Fatal("numbers not equal")
	case slices.Compare(src.SomeList, dst.SomeList) != 0:
		t.Fatal("lists not equal")
	}
}
