package astral

import "io"

var _ Object = (*PtrSpec)(nil)

// PtrSpec describes an optional field. Wire layout: [Bool nil-flag][inner payload if non-nil].
type PtrSpec struct {
	Type String16
}

func (*PtrSpec) ObjectType() string { return "astral.blueprint.ptr_spec" }

func (s *PtrSpec) WriteTo(w io.Writer) (int64, error)  { return Objectify(s).WriteTo(w) }
func (s *PtrSpec) ReadFrom(r io.Reader) (int64, error) { return Objectify(s).ReadFrom(r) }

// ReferencedType satisfies Spec. PtrSpec wraps an optional value of Type; that name must be
// registered before this Blueprint for closure validation to pass.
func (s *PtrSpec) ReferencedType() string { return s.Type.String() }

func init() { _ = Add(&PtrSpec{}) }
