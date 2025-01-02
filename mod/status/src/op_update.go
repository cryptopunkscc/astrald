package status

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

func (ops *Ops) Update(ctx astral.Context, q shell.Query) (err error) {
	t, err := shell.AcceptTerminal(q)
	if err != nil {
		return err
	}
	defer t.Close()

	return ops.mod.Broadcast()
}
