package astral

import (
	"bytes"
	"errors"
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

// TestInterface_PointerReceiverObjectType — §11.2. PrimitiveSpec has its ObjectType
// defined on the pointer receiver; the interface codec wraps the value as a ptrValue
// with skipNilFlag so the pointer-receiver method stays reachable and the type tag
// round-trips.
func TestInterface_PointerReceiverObjectType(t *testing.T) {
	src := testInterface{
		Nonce: NewNonce(),
		Value: &PrimitiveSpec{PrimitiveType: "uint32"},
	}
	var dst testInterface

	var buf bytes.Buffer
	if _, err := Objectify(&src).WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	if _, err := Objectify(&dst).ReadFrom(&buf); err != nil {
		t.Fatal(err)
	}
	got, ok := dst.Value.(*PrimitiveSpec)
	if !ok {
		t.Fatalf("want *PrimitiveSpec, got %T", dst.Value)
	}
	if got.PrimitiveType != "uint32" {
		t.Fatalf("PrimitiveType: want uint32, got %q", got.PrimitiveType)
	}
}

// TestInterface_StringAliasSpecialCase — §11.3. The kind-switch in interfaceValue.WriteTo
// handles String*/Bytes* aliases specially: their ObjectType resolves through the value's
// own method set rather than through ptrValue or objectify(p.Elem()). Test interface
// holding a String32 (passed as a pointer because Object's method set requires it).
func TestInterface_StringAliasSpecialCase(t *testing.T) {
	src := testInterface{
		Nonce: NewNonce(),
		Value: NewString32("special-case"),
	}
	var dst testInterface

	var buf bytes.Buffer
	if _, err := Objectify(&src).WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	if _, err := Objectify(&dst).ReadFrom(&buf); err != nil {
		t.Fatal(err)
	}
	got, ok := dst.Value.(*String32)
	if !ok {
		t.Fatalf("want *String32, got %T", dst.Value)
	}
	if *got != "special-case" {
		t.Fatalf("value: want \"special-case\", got %q", *got)
	}
}

// TestInterface_UnknownNestedType — §11.4. Receiving an interface tag whose type name is
// not registered must surface ErrStreamCorrupted wrapping ErrBlueprintNotFound — silent
// desync would let the rest of the stream misalign.
func TestInterface_UnknownNestedType(t *testing.T) {
	var dst testInterface
	// Craft wire: Nonce (8 bytes) then interface field. Interface field is:
	//   String8 tag (1 byte length + ASCII bytes) followed by the object payload.
	// Use a type name guaranteed not registered.
	bogus := "no.such.type.registered"
	var buf bytes.Buffer

	// Nonce: 8 zero bytes.
	buf.Write(make([]byte, 8))
	// Interface tag: String8 length + name.
	buf.WriteByte(byte(len(bogus)))
	buf.WriteString(bogus)

	_, err := Objectify(&dst).ReadFrom(&buf)
	if err == nil {
		t.Fatal("want error for unknown interface type, got nil")
	}
	if !errors.Is(err, ErrStreamCorrupted) {
		t.Fatalf("want ErrStreamCorrupted wrap, got %v", err)
	}
	if !errors.Is(err, ErrBlueprintNotFound) {
		t.Fatalf("want ErrBlueprintNotFound wrap, got %v", err)
	}
}

// TestInterface_UnmarshalJSON_MissingType — §11.7. Envelope {"Object":null} has an
// empty Type tag; New("") returns nil, so the decoder must error (it cannot guess the
// concrete type from no tag).
func TestInterface_UnmarshalJSON_MissingType(t *testing.T) {
	var dst testInterface
	o := Objectify(&dst)
	// Outer JSON: {"Nonce":"...","Value":{"Object":null}}.
	// Construct a Value field with Object=null and no Type.
	if err := o.UnmarshalJSON([]byte(`{"Nonce":"AAAAAAAAAAA=","Value":{"Object":null}}`)); err == nil {
		t.Fatal("want error for interface JSON with missing Type, got nil")
	}
}

func TestInterface_MarshalJSON_UnparsedObject(t *testing.T) {
	src := testInterface{
		Nonce: NewNonce(),
		Value: NewUnparsedObject("some.unknown.type", []byte{0x01, 0x02}),
	}
	srcObject := Objectify(&src)

	_, err := srcObject.MarshalJSON()
	if err == nil {
		t.Fatal("want error marshaling UnparsedObject through interface field, got nil")
	}
}
