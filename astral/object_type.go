package astral

import (
	"io"
)

// ObjectType is an object containing the object type. ObjectType is an Object itself.
type ObjectType string

// astral:blueprint-ignore
func (*ObjectType) ObjectType() string { return "object_type" }

func (h ObjectType) WriteTo(w io.Writer) (n int64, err error) {
	return String8(h).WriteTo(w)
}

func (h *ObjectType) ReadFrom(r io.Reader) (n int64, err error) {
	return (*String8)(h).ReadFrom(r)
}

func (h ObjectType) String() string { return string(h) }

func init() {
	_ = DefaultBlueprints.Add((*ObjectType)(nil))
}

type TypeWriter func(io.Writer, ObjectType) (int64, error)

var _ TypeWriter = WriteCanonicalType

// WriteCanonicalType writes the object type to the writer in its canonical form.
func WriteCanonicalType(w io.Writer, t ObjectType) (n int64, err error) {
	var m int64

	m, err = Stamp{}.WriteTo(w)
	n = +m
	if err != nil {
		return
	}

	m, err = t.WriteTo(w)
	n = +m
	return
}

var _ TypeWriter = WriteShortType

// WriteShortType writes the object type to the writer in its short form.
func WriteShortType(w io.Writer, t ObjectType) (int64, error) {
	return t.WriteTo(w)
}

type TypeReader func(io.Reader) (ObjectType, int64, error)

var _ TypeReader = ReadCanonicalType

// ReadCanonicalType reads the object type from the reader in its canonical form.
func ReadCanonicalType(r io.Reader) (t ObjectType, n int64, err error) {
	var m int64
	m, err = (&Stamp{}).ReadFrom(r)
	n += m
	if err != nil {
		return
	}

	m, err = t.ReadFrom(r)
	n += m
	return
}

var _ TypeReader = ReadShortType

// ReadShortType reads the object type from the reader in its short form.
func ReadShortType(r io.Reader) (t ObjectType, n int64, err error) {
	n, err = t.ReadFrom(r)
	return
}
