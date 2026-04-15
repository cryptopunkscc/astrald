package tcp

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
)

type opNewEphemeralListenerArgs struct {
	Port astral.Uint16
	In   string `query:"optional"`
	Out  string `query:"optional"`
}

func (mod *Module) OpNewEphemeralListener(ctx *astral.Context, q *routing.IncomingQuery, args opNewEphemeralListenerArgs) (err error) {
	ch := channel.New(q.AcceptRaw(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	err = mod.CreateEphemeralListener(ctx, args.Port, mod.acceptAll)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return ch.Send(&astral.Ack{})
}
