package dir

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

// AliasMap maps aliases to identities.
type AliasMap struct {
	Aliases map[string]*astral.Identity
}

func (a AliasMap) ObjectType() string {
	return "mod.dir.alias_map"
}

func (a AliasMap) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Objectify(&a).WriteTo(w)
}

func (a *AliasMap) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Objectify(a).ReadFrom(r)
}

func (a AliasMap) MarshalJSON() ([]byte, error) {
	return astral.Objectify(&a).MarshalJSON()
}

func (a *AliasMap) UnmarshalJSON(bytes []byte) error {
	return astral.Objectify(a).UnmarshalJSON(bytes)
}

func init() {
	astral.DefaultBlueprints.Add(&AliasMap{})
}
