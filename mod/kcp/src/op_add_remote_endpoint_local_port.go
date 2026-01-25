package kcp

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/kcp"
)

type opAddRemoteEndpointLocalPort struct {
	Endpoint  string
	LocalPort astral.Uint16
	Replace   astral.Bool `query:"optional"`
	In        string      `query:"optional"`
	Out       string      `query:"optional"`
}

func (mod *Module) OpAddRemoteEndpointLocalPort(ctx *astral.Context, q *ops.Query, args opAddRemoteEndpointLocalPort) (err error) {
	endpoint, err := kcp.ParseEndpoint(args.Endpoint)
	if err != nil {
		return q.RejectWithCode(4)
	}

	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	err = mod.SetEndpointLocalSocket(*endpoint, args.LocalPort, args.Replace)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return ch.Send(&astral.Ack{})
}
