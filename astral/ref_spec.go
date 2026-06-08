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

// ReferencedType satisfies Spec. RefSpec inlines another registered Object identified by Type;
// that name must be registered before this Blueprint for closure validation to pass.
func (s *RefSpec) ReferencedType() string { return s.Type.String() }

func init() { _ = Add(&RefSpec{}) }
