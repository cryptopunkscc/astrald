package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
)

type opListSiblingsArgs struct {
	Out  string      `query:"optional"`
	Zone astral.Zone `query:"optional"`
}

func (mod *Module) OpListSiblings(ctx *astral.Context, q *ops.Query, args opListSiblingsArgs) (err error) {
	ctx, cancel := ctx.WithIdentity(q.Caller()).IncludeZone(args.Zone).WithCancel()
	defer cancel()

	ch := channel.New(q.Accept(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	for _, id := range mod.getLinkedSibs() {
		err = ch.Send(id)
		if err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}
	}

	return ch.Send(&astral.EOS{})
}
