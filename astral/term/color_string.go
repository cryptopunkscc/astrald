package term

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

var _ astral.Object = &ColorString{}
var _ PrinterTo = &ColorString{}

type ColorString struct {
	Color astral.String8
	Text  astral.String32
}

func (ColorString) ObjectType() string {
	return "astrald.term.color_string"
}

func (c ColorString) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(c).WriteTo(w)
}

func (c *ColorString) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(c).ReadFrom(r)
}

func (c *ColorString) PrintTo(printer Printer) error {
	return errors.Join(
		printer.Print(&SetColor{c.Color}),
		printer.Print(&c.Text),
		printer.Print(&SetColor{DefaultColor}),
	)
}
