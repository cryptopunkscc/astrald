package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
)

type opRegisterBlueprintArgs struct {
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpRegisterBlueprint(ctx *astral.Context, q *routing.IncomingQuery, args opRegisterBlueprintArgs) error {
	ch := channel.New(q.AcceptRaw(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	return ch.Collect(func(object astral.Object) error {
		bp, ok := object.(*astral.Blueprint)
		if !ok {
			return ch.Send(astral.NewError("expected astral.Blueprint"))
		}

		id, err := mod.RegisterBlueprint(bp)
		if err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}

		return ch.Send(id)
	})
}
