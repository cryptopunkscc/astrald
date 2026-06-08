package astral

import (
	"bytes"
	"io"
	"testing"
)

type testEntry struct {
	Name    String8
	Objects []Object
	Map     map[string]Object
}

// astral:blueprint-ignore
func (e testEntry) ObjectType() string { return "entry" }

func (e testEntry) WriteTo(w io.Writer) (n int64, err error) {
	return Objectify(&e).WriteTo(w)
}

func (e *testEntry) ReadFrom(r io.Reader) (n int64, err error) {
	return Objectify(e).ReadFrom(r)
}

type testComplex struct {
	Entries []testEntry
}

// astral:blueprint-ignore
func (testComplex) ObjectType() string { return "list" }

func (c testComplex) WriteTo(w io.Writer) (n int64, err error) {
	return Objectify(&c).WriteTo(w)
}

func (c *testComplex) ReadFrom(r io.Reader) (n int64, err error) {
	return Objectify(c).ReadFrom(r)
}

func TestObjectifyComplex(t *testing.T) {
	var src = testComplex{
		Entries: []testEntry{
			{
				Name:    "hello",
				Objects: []Object{GenerateIdentity()},
				Map: map[string]Object{
					"foo": GenerateIdentity(),
				},
			},
		},
	}

	var encoded = &bytes.Buffer{}

	// test binary encoding
	_, err := Objectify(&src).WriteTo(encoded)
	if err != nil {
		t.Fatal(err)
	}

	var dst testComplex
	_, err = Objectify(&dst).ReadFrom(bytes.NewReader(encoded.Bytes()))
	if err != nil {
		t.Fatal(err)
	}

	switch {
	case len(dst.Entries) != len(src.Entries):
		t.Fatal("invalid number of entries")
	case dst.Entries[0].Name != "hello":
		t.Fatal("invalid entry name")
	case len(dst.Entries[0].Objects) != len(src.Entries[0].Objects):
		t.Fatal("invalid number of objects")
	case src.Entries[0].Objects[0].ObjectType() != dst.Entries[0].Objects[0].ObjectType():
		t.Fatal("mismatched object type")
	case len(src.Entries[0].Map) != len(dst.Entries[0].Map):
		t.Fatal("invalid map length")
	case dst.Entries[0].Map["foo"].ObjectType() != src.Entries[0].Map["foo"].ObjectType():
		t.Fatal("mismatched map value type")
	}

	// test json encoding
	jdata, err := Objectify(&src).MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	dst = testComplex{}
	err = Objectify(&dst).UnmarshalJSON(jdata)
	if err != nil {
		t.Fatal(err)
	}

	switch {
	case len(dst.Entries) != len(src.Entries):
		t.Fatal("invalid number of entries")
	case dst.Entries[0].Name != "hello":
		t.Fatal("invalid entry name")
	case len(dst.Entries[0].Objects) != len(src.Entries[0].Objects):
		t.Fatal("invalid number of objects")
	case src.Entries[0].Objects[0].ObjectType() != dst.Entries[0].Objects[0].ObjectType():
		t.Fatal("mismatched object type")
	case len(src.Entries[0].Map) != len(dst.Entries[0].Map):
		t.Fatal("invalid map length")
	case dst.Entries[0].Map["foo"].ObjectType() != src.Entries[0].Map["foo"].ObjectType():
		t.Fatal("mismatched map value type")
	}
}

func TestObjectify_UnsupportedKind(t *testing.T) {
	var c chan int
	o := Objectify(&c)

	var buf bytes.Buffer

	_, err := o.WriteTo(&buf)
	if err == nil {
		t.Fatal("WriteTo: expected error for chan kind")
	}

	_, err = o.ReadFrom(&buf)
	if err == nil {
		t.Fatal("ReadFrom: expected error for chan kind")
	}

	_, err = o.MarshalJSON()
	if err == nil {
		t.Fatal("MarshalJSON: expected error for chan kind")
	}

	err = o.UnmarshalJSON([]byte("null"))
	if err == nil {
		t.Fatal("UnmarshalJSON: expected error for chan kind")
	}
}

// TestObjectify_UnsupportedKinds_Sweep — §13.1. Each unsupported Go kind must produce an
// error on every codec, not a silent zero round-trip.
func TestObjectify_UnsupportedKinds_Sweep(t *testing.T) {
	cases := []struct {
		name string
		mk   func() any
	}{
		{"chan", func() any { var c chan int; return &c }},
		{"func", func() any { var f func(); return &f }},
		{"complex64", func() any { var c complex64; return &c }},
		{"complex128", func() any { var c complex128; return &c }},
		{"uintptr", func() any { var u uintptr; return &u }},
		// why: platform-width int/uint are rejected so the wire bytes stay portable across
		// architectures — see the rejection in objectify.go and the prior art in
		// supportedMapKey (map_value.go).
		{"int", func() any { var i int; return &i }},
		{"uint", func() any { var u uint; return &u }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			o := Objectify(c.mk())

			var buf bytes.Buffer
			if _, err := o.WriteTo(&buf); err == nil {
				t.Fatalf("WriteTo: want error for %s, got nil", c.name)
			}
			if _, err := o.ReadFrom(&buf); err == nil {
				t.Fatalf("ReadFrom: want error for %s, got nil", c.name)
			}
			if _, err := o.MarshalJSON(); err == nil {
				t.Fatalf("MarshalJSON: want error for %s, got nil", c.name)
			}
			if err := o.UnmarshalJSON([]byte("null")); err == nil {
				t.Fatalf("UnmarshalJSON: want error for %s, got nil", c.name)
			}
		})
	}
}

func TestObjectify_UnsupportedMapKey(t *testing.T) {
	src := map[float64]string{1.5: "x"}
	srcObj := Objectify(&src)

	var buf bytes.Buffer

	_, err := srcObj.WriteTo(&buf)
	if err == nil {
		t.Fatal("WriteTo: expected error for float64 map key")
	}

	dst := map[float64]string{}
	dstObj := Objectify(&dst)

	_, err = dstObj.ReadFrom(&buf)
	if err == nil {
		t.Fatal("ReadFrom: expected error for float64 map key")
	}

	_, err = srcObj.MarshalJSON()
	if err == nil {
		t.Fatal("MarshalJSON: expected error for float64 map key")
	}

	err = dstObj.UnmarshalJSON([]byte("{}"))
	if err == nil {
		t.Fatal("UnmarshalJSON: expected error for float64 map key")
	}
}

func primitiveRoundTrip[T comparable](t *testing.T, src T) {
	t.Helper()

	var buf bytes.Buffer

	_, err := Objectify(&src).WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	var dst T
	_, err = Objectify(&dst).ReadFrom(&buf)
	if err != nil {
		t.Fatalf("ReadFrom: %v", err)
	}

	if dst != src {
		t.Fatalf("binary round-trip mismatch: got %v, want %v", dst, src)
	}

	jdata, err := Objectify(&src).MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}

	var jdst T
	err = Objectify(&jdst).UnmarshalJSON(jdata)
	if err != nil {
		t.Fatalf("UnmarshalJSON: %v", err)
	}

	if jdst != src {
		t.Fatalf("json round-trip mismatch: got %v, want %v", jdst, src)
	}
}

func TestObjectify_PrimitiveRoundTrip(t *testing.T) {
	t.Run("bool", func(t *testing.T) { primitiveRoundTrip(t, true) })
	t.Run("uint8", func(t *testing.T) { primitiveRoundTrip(t, uint8(200)) })
	t.Run("uint16", func(t *testing.T) { primitiveRoundTrip(t, uint16(50_000)) })
	t.Run("uint32", func(t *testing.T) { primitiveRoundTrip(t, uint32(3_000_000_000)) })
	t.Run("uint64", func(t *testing.T) { primitiveRoundTrip(t, uint64(1<<63|0xdead)) })
	t.Run("int8", func(t *testing.T) { primitiveRoundTrip(t, int8(-42)) })
	t.Run("int16", func(t *testing.T) { primitiveRoundTrip(t, int16(-12_345)) })
	t.Run("int32", func(t *testing.T) { primitiveRoundTrip(t, int32(-1_234_567_890)) })
	t.Run("float32", func(t *testing.T) { primitiveRoundTrip(t, float32(3.5)) })
	t.Run("float64", func(t *testing.T) { primitiveRoundTrip(t, float64(2.71828)) })
}
