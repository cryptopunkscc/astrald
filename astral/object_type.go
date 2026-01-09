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
	_ = Add((*ObjectType)(nil))
}
