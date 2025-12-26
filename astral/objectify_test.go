package astral

import (
	"bytes"
	"io"
	"testing"
)

type testEntry struct {
	Name    String8
	Objects []Object
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
	var m = testComplex{
		Entries: []testEntry{
			{
				Name:    "hello",
				Objects: []Object{GenerateIdentity()},
			},
		},
	}

	var encoded = &bytes.Buffer{}

	_, err := Objectify(&m).WriteTo(encoded)
	if err != nil {
		t.Fatal(err)
	}

	var n testComplex
	_, err = Objectify(&n).ReadFrom(bytes.NewReader(encoded.Bytes()))
	if err != nil {
		t.Fatal(err)
	}

	switch {
	case len(n.Entries) != 1:
		t.Fatal("invalid number of entries")
	case n.Entries[0].Name != "hello":
		t.Fatal("invalid entry name")
	case len(n.Entries[0].Objects) != 1:
		t.Fatal("invalid number of objects")
	case m.Entries[0].Objects[0].ObjectType() != n.Entries[0].Objects[0].ObjectType():
		t.Fatal("mismatched object type")
	}
}
