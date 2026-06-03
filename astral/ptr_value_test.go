package astral

import (
	"bytes"
	"reflect"
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

// TestPtr_InvalidPresenceByte — §10.1. Wire format reserves 0x00 (absent) and 0x01
// (present); any other byte must surface an error and not be silently treated as
// "present" or "absent".
func TestPtr_InvalidPresenceByte(t *testing.T) {
	var dst *string
	dstObject := Objectify(&dst)

	buf := bytes.NewReader([]byte{0x02, 'a'})
	_, err := dstObject.ReadFrom(buf)
	if err == nil {
		t.Fatal("want error for invalid nil flag 0x02, got nil")
	}
}

// TestPtr_ToInterface_RoundTrip — §10.2. Pointer-to-interface composes the pointer and
// interface codecs; assert wire and JSON paths both work for a *Object holding a *String16.
func TestPtr_ToInterface_RoundTrip(t *testing.T) {
	src := struct{ P *Object }{}
	v := Object(NewString16("hi"))
	src.P = &v

	var dst struct{ P *Object }

	var buf bytes.Buffer
	if _, err := Objectify(&src).WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	if _, err := Objectify(&dst).ReadFrom(&buf); err != nil {
		t.Fatal(err)
	}
	if dst.P == nil || *dst.P == nil {
		t.Fatal("dst.P is nil")
	}
	got, ok := (*dst.P).(*String16)
	if !ok || *got != "hi" {
		t.Fatalf("want *String16 \"hi\", got %#v", *dst.P)
	}

	// JSON path
	dst = struct{ P *Object }{}
	j, err := Objectify(&src).MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if err := Objectify(&dst).UnmarshalJSON(j); err != nil {
		t.Fatal(err)
	}
	if dst.P == nil || *dst.P == nil {
		t.Fatal("JSON: dst.P is nil")
	}
	if got, ok := (*dst.P).(*String16); !ok || *got != "hi" {
		t.Fatalf("JSON: want *String16 \"hi\", got %#v", *dst.P)
	}
}

// TestPtr_SkipNilFlag — §10.4. skipNilFlag is the internal optimization interfaceValue
// uses to avoid double-emitting a presence byte. White-box test: construct ptrValue
// directly and assert no presence byte appears on the wire.
func TestPtr_SkipNilFlag(t *testing.T) {
	// Present path: should emit only the payload (no leading flag byte).
	s := String16("hi")
	rv := reflect.ValueOf(&s)
	pv := ptrValue{Value: rv, skipNilFlag: true}

	var buf bytes.Buffer
	if _, err := pv.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	// String16 layout: 2-byte length (BE) + payload. "hi" → [0x00, 0x02, 'h', 'i'].
	want := []byte{0x00, 0x02, 'h', 'i'}
	if !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("skipNilFlag=true wrote %x, want %x (no leading presence byte)", buf.Bytes(), want)
	}

	// ReadFrom with skipNilFlag should consume only payload bytes.
	var dst String16
	dpv := ptrValue{Value: reflect.ValueOf(&dst), skipNilFlag: true}
	if _, err := dpv.ReadFrom(bytes.NewReader(want)); err != nil {
		t.Fatal(err)
	}
	if dst != "hi" {
		t.Fatalf("dst: want \"hi\", got %q", dst)
	}

	// Nil path with skipNilFlag: emits 0 bytes.
	var nilPtr *String16
	npv := ptrValue{Value: reflect.ValueOf(&nilPtr).Elem(), skipNilFlag: true}
	var nilBuf bytes.Buffer
	if _, err := npv.WriteTo(&nilBuf); err != nil {
		t.Fatal(err)
	}
	if nilBuf.Len() != 0 {
		t.Fatalf("nil + skipNilFlag should write zero bytes, got %d", nilBuf.Len())
	}
}
