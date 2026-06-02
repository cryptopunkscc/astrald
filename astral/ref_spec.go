package astral

import "io"

var _ Object = (*RefSpec)(nil)

// RefSpec describes a field whose value is another registered Object type, encoded inline (no type tag).
type RefSpec struct {
	Type String16
}

func (*RefSpec) ObjectType() string { return "astral.blueprint.ref_spec" }

func (s *RefSpec) WriteTo(w io.Writer) (int64, error)  { return Objectify(s).WriteTo(w) }
func (s *RefSpec) ReadFrom(r io.Reader) (int64, error) { return Objectify(s).ReadFrom(r) }

func init() { _ = Add(&RefSpec{}) }
