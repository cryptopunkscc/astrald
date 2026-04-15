package gateway

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
	"github.com/cryptopunkscc/astrald/mod/gateway"
)

type opNodeRegisterArgs struct {
	Visibility gateway.Visibility
	In         string `query:"optional"`
	Out        string `query:"optional"`
}

func (mod *Module) OpNodeRegister(
	ctx *astral.Context,
	q *routing.IncomingQuery,
	args opNodeRegisterArgs,
) (err error) {
	ch := channel.New(q.AcceptRaw(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	socket, err := mod.register(ctx, q.Caller(), args.Visibility, "tcp")
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return ch.Send(&socket)
}
