package objects

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type QueryTagMod = astral.String8

const (
	TagModDefault  QueryTagMod = ""
	TagModExclude  QueryTagMod = "EXCLUDE"
	TagModOptional QueryTagMod = "OPTIONAL"
)

var _ astral.Object = &QueryTag{}

type QueryTag struct {
	Name  astral.String8
	Mod   QueryTagMod
	Value astral.String8
}

func (t QueryTag) ObjectType() string { return "objects.query_tag" }

func (t QueryTag) WriteTo(w io.Writer) (int64, error) {
	return astral.Objectify(&t).WriteTo(w)
}

func (t *QueryTag) ReadFrom(r io.Reader) (int64, error) {
	return astral.Objectify(t).ReadFrom(r)
}

func init() {
	astral.Add(&QueryTag{})
}
