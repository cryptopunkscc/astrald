package kos

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opGetArgs struct {
	Key string
	Out string `query:"optional"`
}

func (mod *Module) OpGet(ctx *astral.Context, q shell.Query, args opGetArgs) error {
	typ, payload, err := mod.db.Get(ctx.Identity(), args.Key)
	if err != nil {
		return q.RejectWithCode(8)
	}

	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	unparsed := astral.NewUnparsedObject(typ, payload)
	object, err := mod.Objects.Blueprints().Parse(unparsed)
	if err != nil {
		object = unparsed
	}

	return ch.Write(object)
}
