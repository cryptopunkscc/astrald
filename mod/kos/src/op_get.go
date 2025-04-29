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

	raw := &astral.RawObject{Type: typ, Payload: payload}
	object, err := mod.Objects.Blueprints().Refine(raw)
	if err != nil {
		object = raw
	}

	return ch.Write(object)
}
