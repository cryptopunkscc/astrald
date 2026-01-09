package shell

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
)

type Prompt struct {
	guestID *astral.Identity
	hostID  *astral.Identity
}

func (p Prompt) Render() string {
	return log.DefaultViewer.Render(p.guestID) +
		"@" +
		log.DefaultViewer.Render(p.hostID) +
		"> "
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

func init() {
	_ = astral.Add(&Prompt{})
}
