package kcp

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opCloseEphemeralListenerArgs struct {
	Endpoint string
	In       string `query:"optional"`
	Out      string `query:"optional"`
}

func (mod *Module) OpCloseEphemeralListener(ctx *astral.Context, q shell.Query, args opCloseEphemeralListenerArgs) (err error) {
	listener, ok := mod.ephemeralListeners.Get(args.Endpoint)
	if !ok {
		return q.RejectWithCode(4)
	}

	ch := astral.NewChannelFmt(q.Accept(), args.In, args.Out)
	defer ch.Close()

	err = listener.Close()
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	return ch.Write(&astral.Ack{})
}
