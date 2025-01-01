package term

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"io"
	"strings"
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
	case "default":
		s = "\033[0m"
	case "black":
		s = "\033[30m"
	case "red":
		s = "\033[31m"
	case "green":
		s = "\033[32m"
	case "yellow":
		s = "\033[33m"
	case "blue":
		s = "\033[34m"
	case "magenta":
		s = "\033[35m"
	case "cyan":
		s = "\033[36m"
	case "white":
		s = "\033[37m"

	case "brightblack":
		s = "\033[90m"
	case "brightred":
		s = "\033[91m"
	case "brightgreen":
		s = "\033[92m"
	case "brightyellow":
		s = "\033[93m"
	case "brightblue":
		s = "\033[94m"
	case "brightmagenta":
		s = "\033[95m"
	case "brightcyan":
		s = "\033[96m"
	case "brightwhite":
		s = "\033[97m"
	}

	printer.w.Write([]byte(s))
}
