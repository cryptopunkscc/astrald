package kos

import (
	"encoding/base64"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opSetObjectArgs struct {
	Key     string
	Type    string `query:"optional"`
	Payload string //base64 encoded payload
}

func (mod *Module) OpSetObject(ctx astral.Context, q shell.Query, args opSetObjectArgs) error {
	payload, err := base64.StdEncoding.DecodeString(args.Payload)
	if err != nil {
		q.Reject()
		return nil
	}

	conn, err := q.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()

	err = mod.db.Set(ctx.Identity(), args.Key, args.Type, payload)
	if err != nil {
		mod.log.Errorv(2, "errors setting %v:%v: %v", ctx.Identity(), args.Key, err)
		_, err = astral.Write(conn, astral.NewError(err.Error()), false)
		return err
	}

	_, err = astral.Write(conn, &astral.Ack{}, false)
	return err
}
