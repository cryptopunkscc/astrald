package term

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/astral"
	"io"
	"time"
)

var _ astral.Error = &Error{}
var _ PrinterTo = &Error{}

type Error struct {
	Time  astral.Time
	Items []astral.Object
}

func (Error) ObjectType() string {
	return "astrald.mod.term.error"
}

func (e Error) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(e).WriteTo(w)
}

func (e *Error) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(e).ReadFrom(r)
}

func (e Error) Error() string {
	var buf = &bytes.Buffer{}
	var p = NewBasicPrinter(buf, &DefaultTypeMap)
	p.Mono = true
	for _, i := range e.Items {
		p.Print(i)
	}
	return buf.String()
}

func (e Error) String() string {
	return e.Error()
}

func Errorf(format string, a ...interface{}) astral.Error {
	return &Error{
		Time:  astral.Time(time.Now()),
		Items: Format(format, a...),
	}
}

func (e Error) PrintTo(printer Printer) error {
	return printer.Print(&ColorString{
		Color: "red",
		Text:  astral.String32(e.Error()),
	})
}
