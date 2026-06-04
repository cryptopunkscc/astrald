package astral

import "io"

var _ Object = (*PrimitiveSpec)(nil)

// PrimitiveSpec describes a field whose value is one of the allowlisted primitive astral types.
type PrimitiveSpec struct {
	PrimitiveType String16
}

func (*PrimitiveSpec) ObjectType() string { return "astral.blueprint.primitive_spec" }

func (s *PrimitiveSpec) WriteTo(w io.Writer) (int64, error)  { return Objectify(s).WriteTo(w) }
func (s *PrimitiveSpec) ReadFrom(r io.Reader) (int64, error) { return Objectify(s).ReadFrom(r) }

// primitiveAllowlist bounds PrimitiveSpec to a fixed set of canonical primitives.
var primitiveAllowlist = []string{
	"string8", "string16", "string32", "string64",
	"uint8", "uint16", "uint32", "uint64",
	"bytes8", "bytes16", "bytes32", "bytes64",
	"bool", "time", "identity", "object_id.sha256",
	"nonce64", "duration", "zone",
}

func isAllowedPrimitive(name string) bool {
	for _, p := range primitiveAllowlist {
		if p == name {
			return true
		}
	}
	return false
}

func init() { _ = Add(&PrimitiveSpec{}) }
