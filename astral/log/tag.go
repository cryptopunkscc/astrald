package log

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/term"
	"io"
)

type Tag astral.String8

func (Tag) ObjectType() string { return "astrald.log.tag" }

func (l Tag) WriteTo(w io.Writer) (n int64, err error) {
	return astral.String8(l).WriteTo(w)
}

func (l *Tag) ReadFrom(r io.Reader) (n int64, err error) {
	return (*astral.String8)(l).ReadFrom(r)
}

func (l Tag) PrintTo(p term.Printer) error {
	if len(l) == 0 {
		return term.Printf(p, "[-]")
	}
	return term.Printf(p, "[%v]", string(l))
}

func init() {
	var v Tag
	astral.DefaultBlueprints.Add(&v)
}
