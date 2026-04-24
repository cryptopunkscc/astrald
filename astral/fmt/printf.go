package fmt

import (
	"bytes"
	"io"
	"os"
	"reflect"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/sig"
)

type Printer struct {
	io.Writer
}

var viewBuilders sig.Map[string, ViewBuilder]

type ViewBuilder func(astral.Object) View

func NewPrinter(writer io.Writer) *Printer {
	return &Printer{Writer: writer}
}

var Stdout = NewPrinter(os.Stdout)
var Stderr = NewPrinter(os.Stderr)

type View interface {
	Render() string
}

func (p *Printer) Print(args ...any) (n int, err error) {
	var written int
	for i, a := range args {
		if i > 0 {
			written, err = p.Writer.Write([]byte(" "))
			n += written
			if err != nil {
				return
			}
		}
		written, err = p.Writer.Write([]byte(ViewFor(a).Render()))
		n += written
		if err != nil {
			return
		}
	}
	return
}

func (p *Printer) Println(args ...any) (n int, err error) {
	n, err = p.Print(args...)
	if err != nil {
		return
	}
	var written int
	written, err = p.Writer.Write([]byte("\n"))
	n += written
	return
}

func (p *Printer) Printf(f string, args ...any) (n int, err error) {
	var list = Format(f, args...)
	var m int

	for _, a := range list {
		m, err = p.Writer.Write([]byte(ViewFor(a).Render()))
		n += m
		if err != nil {
			return
		}
	}
	return
}

func (p *Printer) PrintfOld(format string, args ...any) (n int, err error) {
	var token string
	var written int
	for len(format) > 0 {
		token, format = cutToken(format)

		if len(token) <= 1 {
			written, err = p.Writer.Write([]byte(token))
			n += written
			if err != nil {
				return
			}
			continue
		}

		switch token[0] {
		case '%':
			switch token[1] {
			case 'v':
				if len(args) == 0 {
					written, err = p.Writer.Write([]byte("#[err_arg_missing]"))
					n += written
					if err != nil {
						return
					}
				}

				var a any
				a, args = args[0], args[1:]

				written, err = p.Writer.Write([]byte(ViewFor(a).Render()))
				n += written
				if err != nil {
					return
				}

			case 's', 'd':
				if len(args) == 0 {
					written, err = p.Writer.Write([]byte("#[err_arg_missing]"))
					n += written
					if err != nil {
						return
					}
					break
				}

				var a any
				a, args = args[0], args[1:]

				written, err = p.Writer.Write([]byte(astral.Stringify(a)))
				n += written
				if err != nil {
					return
				}

			default:
				written, err = p.Writer.Write([]byte(token))
				n += written
				if err != nil {
					return
				}
			}

		case '\\':
			switch token[1] {
			case 't':
				written, err = p.Writer.Write([]byte{'\t'})
				n += written
				if err != nil {
					return
				}
			case 'n':
				written, err = p.Writer.Write([]byte{'\n'})
				n += written
				if err != nil {
					return
				}
			case 'r':
				written, err = p.Writer.Write([]byte{'\r'})
				n += written
				if err != nil {
					return
				}
			default:
				written, err = p.Writer.Write([]byte(token))
				n += written
				if err != nil {
					return
				}
			}
		default:
			written, err = p.Writer.Write([]byte(token))
			n += written
			if err != nil {
				return
			}
		}
	}
	return
}

func (p *Printer) Sprint(args ...any) string {
	var out = &bytes.Buffer{}
	NewPrinter(out).Print(args...)
	return out.String()
}

func (p *Printer) Sprintf(format string, args ...any) string {
	var out = &bytes.Buffer{}
	NewPrinter(out).Printf(format, args...)
	return out.String()
}

func (p *Printer) Sprintln(args ...any) string {
	var out = &bytes.Buffer{}
	NewPrinter(out).Println(args...)
	return out.String()
}

func Printf(format string, args ...any) (n int, err error) {
	return Stdout.Printf(format, args...)
}

func Println(args ...any) (n int, err error) {
	return Stdout.Println(args...)
}

func Print(args ...any) (n int, err error) {
	return Stdout.Print(args...)
}

func Sprintf(format string, args ...any) string {
	return Stdout.Sprintf(format, args...)
}

func Sprintln(args ...any) string {
	return Stdout.Sprintln(args...)
}

func Sprint(args ...any) string {
	return Stdout.Sprint(args...)
}

func Fprintf(w io.Writer, format string, args ...any) (n int, err error) {
	return NewPrinter(w).Printf(format, args...)
}

func Fprintln(w io.Writer, args ...any) (n int, err error) {
	return NewPrinter(w).Println(args...)
}

func Fprint(args ...any) (n int, err error) {
	return NewPrinter(os.Stdout).Print(args...)
}

func SetView[T astral.Object](fn func(T) View) {
	var zero T

	v := reflect.ValueOf(&zero).Elem()
	if v.Kind() == reflect.Ptr {
		v.Set(reflect.New(v.Type().Elem()))
	}

	viewBuilders.Replace(zero.ObjectType(), func(object astral.Object) View {
		typed, ok := object.(T)
		if !ok {
			return ErrorView{err: "#[invalid_type]"}
		}
		return fn(typed)
	})
}

func ViewFor(a any) (view View) {
	var ok bool

	view, ok = a.(View)
	if ok {
		return
	}

	if o, ok := a.(astral.Object); ok {
		if builder, found := viewBuilders.Get(o.ObjectType()); found {
			return builder(o)
		}
	}

	return stringView(astral.Stringify(a))
}

func cutToken(s string) (token, left string) {
	if len(s) <= 1 {
		return s, ""
	}

	switch s[0] {
	case '\\':
		return s[0:2], s[2:]
	case '%':
		return s[0:2], s[2:]
	default:
		for i := 1; i < len(s); i++ {
			switch s[i] {
			case '\\', '%':
				return s[:i], s[i:]
			}
		}
		return s, ""
	}
}
