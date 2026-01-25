package dir

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
)

type opResolveArgs struct {
	Name string
	Out  string `query:"optional"`
}

func (mod *Module) OpResolve(ctx *astral.Context, q *ops.Query, args opResolveArgs) (err error) {
	id, err := mod.ResolveIdentity(args.Name)
	if err != nil {
		return q.RejectWithCode(8)
	}

	ch := channel.New(q.Accept(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	return ch.Send(id)
}
