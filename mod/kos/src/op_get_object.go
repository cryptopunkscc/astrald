package kos

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opGetObjectArgs struct {
	Key string
}

func (mod *Module) OpGetObject(ctx astral.Context, q shell.Query, args opGetObjectArgs) error {
	typ, payload, err := mod.db.Get(ctx.Identity(), args.Key)
	if err != nil {
		q.Reject()
		return nil
	}

	conn, err := q.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = astral.Write(conn, &astral.RawObject{Type: typ, Payload: payload}, false)

	return nil
}
