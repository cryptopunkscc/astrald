package ip

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opDefaultGatewayArgs struct {
	Out string `query:"optional"`
}

func (mod *Module) OpDefaultGateway(ctx *astral.Context, q shell.Query, args opDefaultGatewayArgs) (err error) {
	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	gw, err := mod.DefaultGateway()
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	return ch.Write(&gw)
}
