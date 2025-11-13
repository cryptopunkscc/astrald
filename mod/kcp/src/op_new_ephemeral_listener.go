package kcp

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/kcp"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opNewEphemeralListenerArgs struct {
	Endpoint string
	In       string `query:"optional"`
	Out      string `query:"optional"`
}

func (mod *Module) OpNewEphemeralListener(ctx *astral.Context, q shell.Query, args opNewEphemeralListenerArgs) (err error) {
	ep, err := kcp.ParseEndpoint(args.Endpoint)
	if err != nil {
		return q.RejectWithCode(4)
	}

	ch := astral.NewChannelFmt(q.Accept(), args.In, args.Out)
	defer ch.Close()

	err = mod.CreateEphemeralListener(*ep)
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	return ch.Write(&astral.Ack{})
}
