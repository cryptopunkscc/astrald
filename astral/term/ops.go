package term

import (
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

type SetColor struct {
	Color astral.String8
}

func (SetColor) ObjectType() string { return "astrald.mod.shell.ops.set_color" }

func (o SetColor) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(o).WriteTo(w)
}

func (o *SetColor) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(o).ReadFrom(r)
}
