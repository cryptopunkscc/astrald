package term

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/sig"
	"os"
	"reflect"
)

var DefaultPrinter *Printer

type ObjectPrinter func(astral.Object, Renderer)
type TypePrinter func(any, Renderer)

type Printer struct {
	Renderer       Renderer
	ObjectPrinters sig.Map[string, ObjectPrinter]
	TypePrinters   sig.Map[string, TypePrinter]
	StringPrinter  func(Renderer, string)
	ErrorPrinter   func(Renderer, error)
	IntPrinter     func(Renderer, int64)
	UintPrinter    func(Renderer, uint64)
	BoolPrinter    func(Renderer, bool)
	FloatPrinter   func(Renderer, float64)
}

type printerTo interface {
	PrintTo(*Printer)
}

func NewPrinter(r Renderer) *Printer {
	return &Printer{Renderer: r}
}

func (p *Printer) Printf(f string, v ...interface{}) {
	p.Print(Format(f, v...)...)
}

func (p *Printer) PrintfTo(r Renderer, f string, v ...interface{}) {
	p.PrintTo(r, Format(f, v...)...)
}

func (p *Printer) Print(v ...interface{}) {
	for _, i := range v {
		p.print(p.Renderer, i)
	}
}

func (p *Printer) PrintTo(r Renderer, v ...interface{}) {
	for _, i := range v {
		p.print(r, i)
	}
}

func (p *Printer) print(r Renderer, v any) {
	if pto, ok := v.(printerTo); ok {
		pto.PrintTo(p)
		return
	}

	if o, ok := v.(astral.Object); ok {
		op, ok := p.ObjectPrinters.Get(o.ObjectType())
		if ok {
			op(o, r)
			return
		}
	}

	switch v := v.(type) {
	case string:
		if p.StringPrinter != nil {
			p.StringPrinter(r, v)
			return
		}

	case int, int8, int16, int32, int64:
		if p.IntPrinter != nil {
			p.IntPrinter(r, v.(int64))
			return
		}

	case uint, uint8, uint16, uint32, uint64:
		if p.UintPrinter != nil {
			p.UintPrinter(r, v.(uint64))
			return
		}

	case float32, float64:
		if p.FloatPrinter != nil {
			p.FloatPrinter(r, v.(float64))
			return
		}

	case bool:
		if p.BoolPrinter != nil {
			p.BoolPrinter(r, v)
			return
		}
	}

	if s, ok := p.TypePrinters.Get(reflect.TypeOf(v).String()); ok {
		s(v, r)
		return
	}

	if s, ok := v.(fmt.Stringer); ok {
		if p.StringPrinter != nil {
			p.StringPrinter(r, s.String())
			return
		}
	}

	if s, ok := v.(error); ok {
		if p.ErrorPrinter != nil {
			p.ErrorPrinter(r, s)
			return
		}
	}

	r.Text(fmt.Sprintf("%v", v))
}

func Printf(f string, v ...interface{}) {
	DefaultPrinter.Printf(f, v...)
}

func PrintfTo(r Renderer, f string, v ...interface{}) {
	DefaultPrinter.PrintfTo(r, f, v...)
}

func Print(v ...interface{}) {
	DefaultPrinter.Print(v...)
}

func PrintTo(r Renderer, v ...interface{}) {
	DefaultPrinter.PrintTo(r, v...)
}

func init() {
	DefaultPrinter = NewPrinter(NewLinuxTerminal(os.Stdout))
}
