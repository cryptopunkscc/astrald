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
