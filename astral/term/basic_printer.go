package term

import (
	"fmt"
	"io"
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
)

var _ Printer = &BasicPrinter{}

type BasicPrinter struct {
	Mono bool
	Translator
	w io.Writer
}

func (printer BasicPrinter) Translate(object astral.Object) astral.Object {
	if printer.Translator == nil {
		return object
	}
	return printer.Translator.Translate(object)
}

func NewBasicPrinter(w io.Writer, m Translator) *BasicPrinter {
	return &BasicPrinter{w: w, Translator: m}
}

func (printer BasicPrinter) Print(objects ...astral.Object) (err error) {
	for _, o := range objects {
		err = printer.print(o)
		if err != nil {
			return
		}
	}
	return
}

func (printer BasicPrinter) print(object astral.Object) (err error) {
	if printer.Translator != nil {
		object = printer.Translator.Translate(object)
	}

	switch object := object.(type) {
	case *SetColor:
		printer.setColor(object.Color.String())
		return
	}

	if strings.HasPrefix(object.ObjectType(), "astrald.mod.shell.ops") {
		// ignore unsupported ops
		return nil
	}

	if p, ok := object.(PrinterTo); ok {
		return p.PrintTo(printer)
	}

	if s, ok := object.(fmt.Stringer); ok {
		_, err = printer.w.Write([]byte(s.String()))
		return
	}

	raw, err := astral.ToRaw(object)
	if err != nil {
		return err
	}

	text, err := raw.MarshalText()
	if err != nil {
		return err
	}

	printer.w.Write(text)

	return nil
}

func (printer BasicPrinter) setColor(color string) {
	if printer.Mono {
		return
	}

	var s string
	switch color {
	case ColorDefault:
		s = "\033[0m"
	case ColorBlack:
		s = "\033[30m"
	case ColorRed:
		s = "\033[31m"
	case ColorGreen:
		s = "\033[32m"
	case ColorYellow:
		s = "\033[33m"
	case ColorBlue:
		s = "\033[34m"
	case ColorMagenta:
		s = "\033[35m"
	case ColorCyan:
		s = "\033[36m"
	case ColorWhite:
		s = "\033[37m"

	case ColorBrightBlack:
		s = "\033[90m"
	case ColorBrightRed:
		s = "\033[91m"
	case ColorBrightGreen:
		s = "\033[92m"
	case ColorBrightYellow:
		s = "\033[93m"
	case ColorBrightBlue:
		s = "\033[94m"
	case ColorBrightMagenta:
		s = "\033[95m"
	case ColorBrightCyan:
		s = "\033[96m"
	case ColorBrightWhite:
		s = "\033[97m"
	}

	printer.w.Write([]byte(s))
}
