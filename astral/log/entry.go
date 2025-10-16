package log

import (
	"io"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/term"
)

var _ astral.Object = &Entry{}
var _ term.PrinterTo = &Entry{}

type Entry struct {
	Origin  *astral.Identity
	Level   Level
	Type    Type
	Time    astral.Time
	Tag     Tag
	Objects []astral.Object
}

func NewEntry(origin *astral.Identity, level Level, Type Type, tag Tag, f string, v ...any) *Entry {
	return &Entry{
		Origin:  origin,
		Level:   level,
		Type:    Type,
		Tag:     tag,
		Time:    astral.Time(time.Now()),
		Objects: term.Format(f, v...),
	}
}

func (Entry) ObjectType() string {
	return "astrald.log.entry"
}

func (e Entry) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(e).WriteTo(w)
}

func (e *Entry) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(e).ReadFrom(r)
}

func (e Entry) PrintTo(p term.Printer) error {
	// set header color
	var c = &term.SetColor{"white"}
	var o = &term.SetColor{"brightblack"}

	// print entry header
	term.Printf(p, "%v%v %v%v %v%v %v%v %v%v ",
		o, &Origin{e.Origin},
		c, e.Level,
		c, Time(e.Time),
		c, e.Type,
		c, e.Tag,
	)

	// print entry objects
	for _, o := range e.Objects {
		p.Print(&term.SetColor{term.ColorDefault})
		p.Print(o)
	}
	return nil
}

func init() {
	astral.DefaultBlueprints.Add(&Entry{})
}
