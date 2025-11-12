package kcp

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/kcp"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opRegisterEphemeralEndpointArgs struct {
	Endpoint  string
	LocalPort uint16

	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpRegisterEphemeralEndpoint(ctx *astral.Context, q shell.Query, args opRegisterEphemeralEndpointArgs) (err error) {
	endpoint, err := kcp.ParseEndpoint(args.Endpoint)
	if err != nil {
		return q.RejectWithCode(4)
	}

	ch := astral.NewChannelFmt(q.Accept(), args.In, args.Out)
	defer ch.Close()

	err = mod.SetEndpointLocalSocket(*endpoint, astral.Uint16(args.LocalPort))
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	return ch.Write(&astral.Ack{})
}
