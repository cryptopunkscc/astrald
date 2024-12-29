package admin

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/astral/term"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"io"
)

const useColorTerminal = true

var _ admin.Terminal = &ColorTerminal{}

type ColorTerminal struct {
	userIdentity *astral.Identity
	color        bool
	log          *log.Logger
	printer      term.Printer
	io.ReadWriter
}

func NewColorTerminal(rw astral.Conn, logger *log.Logger) *ColorTerminal {
	l := logger.Tag("")

	p := term.NewBasicPrinter(rw, &term.DefaultTypeMap)
	p.Mono = !useColorTerminal

	return &ColorTerminal{
		color:        useColorTerminal,
		userIdentity: rw.RemoteIdentity(),
		ReadWriter:   rw,
		printer:      p,
		log:          l,
	}
}

func (t *ColorTerminal) Sprintf(f string, v ...any) string {
	var buf = &bytes.Buffer{}
	var p = term.NewBasicPrinter(buf, &term.DefaultTypeMap)
	p.Mono = !useColorTerminal
	term.Printf(p, f, v...)

	return buf.String()
}

func (t *ColorTerminal) Printf(f string, v ...any) {
	term.Printf(t.printer, f, v...)
}

func (t *ColorTerminal) Println(v ...any) {
	if t.color {
		for _, i := range v {
			var o = term.Objectify(i)
			txt := term.Render(o, &term.DefaultTypeMap, false)
			fmt.Fprintln(t, txt)
		}
		return
	}
	fmt.Fprintln(t, v...)
}

func (t *ColorTerminal) Scanf(f string, v ...any) {
	fmt.Fscanf(t, f, v...)
}

func (t *ColorTerminal) ScanLine() (string, error) {
	var scanner = bufio.NewScanner(t)
	if !scanner.Scan() {
		return "", io.EOF
	}
	return scanner.Text(), nil
}

func (t *ColorTerminal) Color() bool {
	return t.color
}

func (t *ColorTerminal) SetColor(Color bool) {
	t.color = Color
}

func (t *ColorTerminal) UserIdentity() *astral.Identity {
	return t.userIdentity
}

func (t *ColorTerminal) SetUserIdentity(identity *astral.Identity) {
	t.userIdentity = identity
}
