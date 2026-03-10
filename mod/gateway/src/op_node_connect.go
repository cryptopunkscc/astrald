package gateway

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
)

type opNodeConnectArgs struct {
	Target *astral.Identity
	In     string `query:"optional"`
	Out    string `query:"optional"`
}

func (mod *Module) OpNodeConnect(
	ctx *astral.Context,
	q *ops.Query,
	args opNodeConnectArgs,
) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	socket, err := mod.connectTo(q.Caller(), args.Target, "tcp")
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return ch.Send(&socket)
}
