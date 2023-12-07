package admin

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/modules"
	"io"
)

const ModuleName = "admin"

type API interface {
	AddCommand(name string, cmd Command) error
}

type Command interface {
	Exec(out Terminal, args []string) error
}

type Terminal interface {
	UserIdentity() id.Identity
	Sprintf(f string, v ...any) string
	Printf(f string, v ...any)
	Println(v ...any)
	Scanf(f string, v ...any)
	ScanLine() (string, error)
	Color() bool
	SetColor(bool)
	io.Writer
}

// Format types are used to format output text on the terminal. Example:
// term.Println("normal %s %s", Keyword("keyword"), Faded("faded"))

type Header string
type Keyword string
type Faded string
type Important string

func Load(node modules.Node) (API, error) {
	api, ok := node.Modules().Find(ModuleName).(API)
	if !ok {
		return nil, modules.ErrNotFound
	}
	return api, nil
}
