package astral

import (
	"bytes"
	"testing"
)

type testMap struct {
	SomeMap map[Uint8]String8
}

func TestMapWithValues(t *testing.T) {
	var err error
	var src, dst testMap

	src.SomeMap = make(map[Uint8]String8)
	src.SomeMap[1] = "hello world"

	srcObject := Objectify(&src)

	// test binary marshaling
	var buf = &bytes.Buffer{}
	_, err = srcObject.WriteTo(buf)
	if err != nil {
		t.Fatal(err)
	}

	dst.SomeMap = nil
	dstObject := Objectify(&dst)
	_, err = dstObject.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}

	switch {
	case len(dst.SomeMap) != len(src.SomeMap):
		t.Fatal("map lengths not equal")
	case dst.SomeMap[1] != src.SomeMap[1]:
		t.Fatal("map values not equal")
	}
}

func TestMapEmpty(t *testing.T) {
	var err error
	var src, dst testMap

	srcObject := Objectify(&src)

	// test binary marshaling
	var buf = &bytes.Buffer{}
	_, err = srcObject.WriteTo(buf)
	if err != nil {
		t.Fatal(err)
	}

	dst.SomeMap = make(map[Uint8]String8)
	dst.SomeMap[1] = "hello world"

	dstObject := Objectify(&dst)
	_, err = dstObject.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}

	if len(dst.SomeMap) != 0 {
		t.Fatal("interface values not equal")
	}

	// test json marshaling
	//dst.SomeMap = make(map[int]string)
	//dst.SomeMap[1] = "hello world"
	//
	//jdata, err := srcObject.MarshalJSON()
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//dstObject = Objectify(&dst)
	//err = dstObject.UnmarshalJSON(jdata)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//if len(dst.SomeMap) != 0 {
	//	t.Fatal("interface values not equal")
	//}
}
