package status

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

func (ops *Ops) Scan(ctx *astral.Context, q shell.Query) (err error) {
	q.Accept().Close()

	return ops.mod.Scan()
}
