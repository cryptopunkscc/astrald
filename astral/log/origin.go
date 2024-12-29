package log

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"strings"
)
import "github.com/cryptopunkscc/astrald/astral/term"

type Origin struct {
	Identity *astral.Identity
}

func (o Origin) PrintTo(p term.Printer) error {
	var id = astral.String(term.Render(
		o.Identity,
		p.(term.Translator),
		true,
	))

	if strings.Compare(string(id), o.Identity.String()) == 0 {
		id = astral.String(o.Identity.Fingerprint())
	}

	var s = astral.String(fmt.Sprintf("[%s]", &id))

	return p.Print(&s)
}
