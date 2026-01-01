package log

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
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

func (l Type) String() string {
	switch l {
	case 0:
		return "-"
	case 1:
		return "I"
	case 2:
		return "E"
	default:
		return "?"
	}
}

func init() {
	var t Type
	astral.DefaultBlueprints.Add(&t)
}
