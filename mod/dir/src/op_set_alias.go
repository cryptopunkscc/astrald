package dir

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
)

type opSetAliasArgs struct {
	ID    *astral.Identity
	Alias string
	Out   string `query:"optional"`
}

func (mod *Module) OpSetAlias(ctx *astral.Context, q *ops.Query, args opSetAliasArgs) (err error) {
	ch := q.AcceptChannel(channel.WithOutputFormat(args.Out))
	defer ch.Close()

	err = mod.SetAlias(args.ID, args.Alias)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	return ch.Send(&astral.Ack{})
}
