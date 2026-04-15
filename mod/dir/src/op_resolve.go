package dir

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
)

type opResolveArgs struct {
	Name string
	Out  string `query:"optional"`
}

func (mod *Module) OpResolve(ctx *astral.Context, q *routing.IncomingQuery, args opResolveArgs) (err error) {
	ch := q.Accept(channel.WithOutputFormat(args.Out))
	defer ch.Close()

	id, err := mod.ResolveIdentity(args.Name)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	return ch.Send(id)
}
