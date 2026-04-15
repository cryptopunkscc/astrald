package objects

import (
	"slices"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
)

type opTypesArgs struct {
	Out string `query:"optional"`
}

func (mod *Module) OpTypes(ctx *astral.Context, q *routing.IncomingQuery, args opTypesArgs) (err error) {
	ch := channel.New(q.AcceptRaw(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	types := astral.DefaultBlueprints().Types()
	slices.Sort(types)

	for _, name := range types {
		err = ch.Send((*astral.String8)(&name))
		if err != nil {
			return
		}
	}

	return
}
