package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/user"
)

type opFooArgs struct {
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpFoo(
	ctx *astral.Context,
	q *ops.Query,
	args opFooArgs,
) (err error) {

	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	authorized := mod.Auth.Authorize(ctx, &user.RevokeContractAction{
		Action: auth.NewAction(q.Caller()),
	})
	if authorized {
		trueVal := astral.Bool(false)
		return ch.Send(&trueVal)
	}

	falseVal := astral.Bool(false)
	return ch.Send(&falseVal)
}
