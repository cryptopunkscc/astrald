package gateway

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/gateway"
)

type opNodeBindArgs struct {
	Visibility gateway.Visibility
	In         string `query:"optional"`
	Out        string `query:"optional"`
}

func (mod *Module) OpNodeBind(
	ctx *astral.Context,
	q *ops.Query,
	args opNodeBindArgs,
) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	socket, err := mod.bind(ctx, q.Caller(), args.Visibility, "tcp")
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return ch.Send(&socket)
}
