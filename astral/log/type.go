package log

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/term"
	"io"
)

type Type astral.Uint8

const (
	Normal = Type(iota)
	Info
	Error
)

func (Type) ObjectType() string { return "astrald.log.type" }

func (l Type) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Uint8(l).WriteTo(w)
}

func (l *Type) ReadFrom(r io.Reader) (n int64, err error) {
	return (*astral.Uint8)(l).ReadFrom(r)
}

func (l Type) PrintTo(p term.Printer) error {
	var s = "?"
	var c = astral.String8("")
	switch l {
	case 0:
		s, c = "-", "default"
	case 1:
		s, c = "I", "green"
	case 2:
		s, c = "E", "red"
	}
	return term.Printf(p, "%v%v%v",
		&term.SetColor{c},
		s,
		&term.SetColor{"default"},
	)
}

func init() {
	var t Type
	astral.DefaultBlueprints.Add(&t)
}
