package log

import (
	"github.com/cryptopunkscc/astrald/astral"
	print2 "github.com/cryptopunkscc/astrald/astral/term"
	"io"
	"strconv"
)

type Level astral.Uint8

func (Level) ObjectType() string { return "astrald.log.level" }

func (l Level) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Uint8(l).WriteTo(w)
}

func (l *Level) ReadFrom(r io.Reader) (n int64, err error) {
	return (*astral.Uint8)(l).ReadFrom(r)
}

func (l Level) PrintTo(p print2.Printer) error {
	return print2.Printf(p, "(%v)", strconv.Itoa(int(l)))
}

func init() {
	var l Level
	astral.DefaultBlueprints.Add(&l)
}
