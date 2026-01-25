package kos

import (
	"bytes"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
)

type opGetArgs struct {
	Key string
	Out string `query:"optional"`
}

func (mod *Module) OpGet(ctx *astral.Context, q *ops.Query, args opGetArgs) error {
	typ, payload, err := mod.db.Get(ctx.Identity(), args.Key)
	if err != nil {
		return q.RejectWithCode(8)
	}

	ch := channel.New(q.Accept(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	// parse the object if possible, so that we can send json and text objects
	object := astral.New(typ)
	if object != nil {
		_, err = object.ReadFrom(bytes.NewReader(payload))
		if err != nil {
			return err
		}

		return ch.Send(object)
	}

	// try to send an unparsed object
	return ch.Send(astral.NewUnparsedObject(typ, payload))
}
