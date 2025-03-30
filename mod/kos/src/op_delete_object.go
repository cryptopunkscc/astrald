package kos

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opDeleteObjectArgs struct {
	Key string
}

func (mod *Module) OpDeleteObject(ctx astral.Context, q shell.Query, args opDeleteObjectArgs) error {
	conn, err := q.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()

	err = mod.db.Delete(ctx.Identity(), args.Key)
	if err != nil {
		_, err = astral.Write(conn, astral.NewError(err.Error()), false)
		return err
	}

	_, err = astral.Write(conn, &astral.Ack{}, false)
	return err
}
