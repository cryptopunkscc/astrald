package ip

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opDefaultGatewayArgs struct {
	Out string `query:"optional"`
}

func (mod *Module) OpDefaultGateway(ctx *astral.Context, q shell.Query, args opDefaultGatewayArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	gw, err := mod.DefaultGateway()
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return ch.Send(&gw)
}
