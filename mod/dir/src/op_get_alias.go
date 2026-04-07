package dir

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
)

type opGetAliasArgs struct {
	ID  *astral.Identity
	Out string `query:"optional"`
}

func (mod *Module) OpGetAlias(ctx *astral.Context, q *ops.Query, args opGetAliasArgs) (err error) {
	ch := q.AcceptChannel(channel.WithOutputFormat(args.Out))
	defer ch.Close()

	alias, err := mod.GetAlias(args.ID)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	return ch.Send((*astral.String8)(&alias))
}
