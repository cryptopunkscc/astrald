package astral

import "io"

var _ Object = (*ObjectSpec)(nil)

// ObjectSpec describes a field that holds any Object; the value is encoded polymorphically
// (type tag + payload) on the wire.
type ObjectSpec struct{}

func (*ObjectSpec) ObjectType() string { return "astral.blueprint.object_spec" }

func (*ObjectSpec) WriteTo(io.Writer) (int64, error)  { return 0, nil }
func (*ObjectSpec) ReadFrom(io.Reader) (int64, error) { return 0, nil }

func init() { _ = Add(&ObjectSpec{}) }
