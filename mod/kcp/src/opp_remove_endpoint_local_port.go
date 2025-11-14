package kcp

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/kcp"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opRemoveRemoteEndpointLocalPort struct {
	Endpoint astral.String
	In       string `query:"optional"`
	Out      string `query:"optional"`
}

func (mod *Module) OpRemoveRemoteEndpointLocalPort(ctx *astral.Context, q shell.Query, args opRemoveRemoteEndpointLocalPort) (err error) {
	endpoint, err := kcp.ParseEndpoint(string(args.Endpoint))
	if err != nil {
		return q.RejectWithCode(4)
	}

	ch := astral.NewChannelFmt(q.Accept(), args.In, args.Out)
	defer ch.Close()

	err = mod.RemoveEndpointLocalSocket(*endpoint)
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	return ch.Write(&astral.Ack{})
}
