package admin

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/term"
	"io"
	"strings"
)

const ModuleName = "admin"
const ActionAccess = "mod.admin.access"
const ActionSudo = "mod.admin.sudo"

type Module interface {
	AddCommand(name string, cmd Command) error
}

type Command interface {
	Exec(term Terminal, args []string) error
}

type Terminal interface {
	UserIdentity() *astral.Identity
	SetUserIdentity(*astral.Identity)
	Sprintf(f string, v ...any) string
	Printf(f string, v ...any)
	Println(v ...any)
	Scanf(f string, v ...any)
	ScanLine() (string, error)
	Color() bool
	SetColor(bool)
	io.Writer
}

// Formatting types are used to format output text on the terminal. Example:
// term.Println("normal %v %v", Keyword("keyword"), Faded("faded"))

type Header string

func (h Header) PrintTo(printer term.Printer) error {
	var s = astral.String(strings.ToUpper(string(h)))
	return printer.Print(&s)
}

type Keyword string

func (h Keyword) PrintTo(printer term.Printer) error {
	var s = astral.String(h)
	return printer.Print(
		&term.SetColor{term.HighlightColor},
		&s,
		&term.SetColor{term.DefaultColor},
	)
}

type Faded string

func (h Faded) PrintTo(printer term.Printer) error {
	var s = astral.String(h)
	return printer.Print(
		&term.SetColor{"white"},
		&s,
		&term.SetColor{term.DefaultColor},
	)
}

type Important string

func (h Important) PrintTo(printer term.Printer) error {
	var s = astral.String(h)
	return printer.Print(
		&term.SetColor{term.HighlightColor},
		&s,
		&term.SetColor{term.DefaultColor},
	)
}
