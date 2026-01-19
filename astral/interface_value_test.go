package astral

import (
	"bytes"
	"testing"
)

type testInterface struct {
	Nonce Nonce
	Value Object
}

func TestInterfaceWithIdentity(t *testing.T) {
	var err error
	var src, dst testInterface
	var srcID *Identity
	srcID = GenerateIdentity()

	src.Nonce = NewNonce()
	src.Value = srcID

	srcObject := Objectify(&src)

	// test binary marshaling
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

	if dst.Value == nil {
		t.Fatal("interface values not equal")
	}

	// test json marshaling
	dst.Value = nil

	jdata, err := srcObject.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	dstObject = Objectify(&dst)
	err = dstObject.UnmarshalJSON(jdata)
	if err != nil {
		t.Fatal(err)
	}

	if dst.Nonce != src.Nonce {
		t.Fatal("nonce values differ")
	}

	if dst.Value == nil {
		t.Fatal("interface values not equal")
	}

	dstID, ok := dst.Value.(*Identity)
	if !ok {
		t.Fatal("invalid object type")
	}

	if !dstID.IsEqual(srcID) {
		t.Fatal("interface values not equal")
	}
}

func TestInterfaceNil(t *testing.T) {
	var err error
	var src, dst testInterface
	src.Value = nil

	srcObject := Objectify(&src)

	// test binary marshaling
	var buf = &bytes.Buffer{}
	_, err = srcObject.WriteTo(buf)
	if err != nil {
		t.Fatal(err)
	}

	dst.Value = NewString32("hello world")

	dstObject := Objectify(&dst)
	_, err = dstObject.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}

	if dst.Value != nil {
		t.Fatal("interface values not equal")
	}

	// test json marshaling
	dst.Value = NewString32("hello world")

	jdata, err := srcObject.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	dstObject = Objectify(&dst)
	err = dstObject.UnmarshalJSON(jdata)
	if err != nil {
		t.Fatal(err)
	}

	if dst.Value != nil {
		t.Fatal("interface values not equal")
	}
}
