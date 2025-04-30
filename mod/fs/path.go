package fs

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/term"
	"io"
	"strings"
)

type Path string

// astral

func (Path) ObjectType() string {
	return "mod.fs.path"
}

func (p Path) WriteTo(w io.Writer) (n int64, err error) {
	return astral.String16(p).WriteTo(w)
}

func (p *Path) ReadFrom(r io.Reader) (n int64, err error) {
	return (*astral.String16)(p).ReadFrom(r)
}

// json

// text

func (p Path) MarshalText() (text []byte, err error) {
	return []byte(p), nil
}

func (p *Path) UnmarshalText(text []byte) error {
	*p = Path(text)
	return nil
}

// other

func (p Path) String() string {
	return string(p)
}

func (p Path) PrintTo(printer term.Printer) error {
	var list []astral.Object
	var sep = astral.String16("/")

	for _, s := range strings.Split(string(p), "/")[1:] {
		if len(s) == 0 {
			continue
		}
		var s = astral.String(s)
		list = append(list,
			&term.SetColor{Color: "white"},
			&sep,
			&term.SetColor{Color: term.HighlightColor},
			&s,
			&term.SetColor{Color: term.DefaultColor},
		)
	}

	printer.Print(list...)

	return nil
}

func init() {
	var p Path
	astral.DefaultBlueprints.Add(&p)
}
