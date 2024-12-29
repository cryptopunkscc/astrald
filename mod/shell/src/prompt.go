package shell

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/term"
	"io"
)

type Prompt struct {
	guestID *astral.Identity
	hostID  *astral.Identity
}

func (p Prompt) PrintTo(printer term.Printer) error {
	term.Printf(printer, "%v@%v> ", p.guestID, p.hostID)
	return nil
}

func (Prompt) ObjectType() string {
	return ""
}

func (p Prompt) WriteTo(w io.Writer) (n int64, err error) {
	return 0, nil
}

func (p Prompt) ReadFrom(r io.Reader) (n int64, err error) {
	return 0, nil
}
