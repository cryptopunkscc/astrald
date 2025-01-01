package shell

import (
	"bufio"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/term"
	"io"
)

var _ term.Printer = &Terminal{}

type Terminal struct {
	printer term.Printer
	scanner *bufio.Scanner
	rw      io.ReadWriter
}

func NewTerminal(rw io.ReadWriter) *Terminal {
	return &Terminal{
		rw:      rw,
		printer: term.NewBasicPrinter(rw, &term.DefaultTypeMap),
		scanner: bufio.NewScanner(rw),
	}
}

func (e *Terminal) ReadLine() (line string, err error) {
	if e.scanner.Scan() {
		line = e.scanner.Text()
		return
	}
	return "", e.scanner.Err()
}

func (e *Terminal) Write(p []byte) (n int, err error) {
	return e.rw.Write(p)
}

func (e *Terminal) Close() error {
	if c, ok := e.rw.(io.Closer); ok {
		return c.Close()
	}
	return errors.New("rw does not implement io.Closer")
}

func (e *Terminal) Printf(f string, v ...interface{}) error {
	return term.Printf(e, f, v...)
}

func (e *Terminal) Print(objects ...astral.Object) (err error) {
	return e.printer.Print(objects...)
}
