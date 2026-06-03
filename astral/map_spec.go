package astral

import "io"

var _ Object = (*MapSpec)(nil)

// MapSpec describes a field whose value is a StringMap or IntMap, depending on KeyType. An empty
// ValueType means heterogeneous values (each carries its own type tag).
type MapSpec struct {
	KeyType   String16
	ValueType String16
}

func (*MapSpec) ObjectType() string { return "astral.blueprint.map_spec" }

func (s *MapSpec) WriteTo(w io.Writer) (int64, error)  { return Objectify(s).WriteTo(w) }
func (s *MapSpec) ReadFrom(r io.Reader) (int64, error) { return Objectify(s).ReadFrom(r) }

// mapKeyAllowlist bounds MapSpec.KeyType to the supported RuntimeMap key shapes: string16
// (string-keyed) and fixed-width unsigned integers 1/2/4/8 bytes.
var mapKeyAllowlist = []string{
	"string16",
	"uint8", "uint16", "uint32", "uint64",
}

func isAllowedMapKey(name string) bool {
	for _, k := range mapKeyAllowlist {
		if k == name {
			return true
		}
	}
	return false
}

func init() { _ = Add(&MapSpec{}) }
