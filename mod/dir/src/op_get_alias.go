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
	ch := channel.New(q.Accept(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	alias, err := mod.GetAlias(args.ID)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return ch.Send((*astral.String8)(&alias))
}
