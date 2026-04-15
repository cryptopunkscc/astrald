package dir

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
)

type opAliasMapArgs struct {
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpAliasMap(ctx *astral.Context, q *routing.IncomingQuery, args opAliasMapArgs) (err error) {
	ch := q.Accept(channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	return ch.Send(mod.AliasMap())
}
