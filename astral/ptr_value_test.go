package astral

import (
	"bytes"
	"strings"
	"testing"
)

func TestPtrString(t *testing.T) {
	var src, dst *string
	src = (*string)(NewString32("hello world"))

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

	if strings.Compare(*dst, *src) != 0 {
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

	if strings.Compare(*dst, *src) != 0 {
		t.Fatal("slice values not equal")
	}
}

func TestPtrNil(t *testing.T) {
	var src, dst *string
	dst = (*string)(NewString32("hello world"))

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

	if dst != nil {
		t.Fatal("binary values not equal")
	}
	dst = (*string)(NewString32("hello world"))

	dstObject = Objectify(&dst)

	// test json marshaling
	json, err := srcObject.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	err = dstObject.UnmarshalJSON(json)
	if err != nil {
		t.Fatal(err)
	}

	if dst != nil {
		t.Fatal("json values not equal")
	}
}
