package kcp

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/kcp"
)

type opRemoveRemoteEndpointLocalPort struct {
	Endpoint astral.String8
	In       string `query:"optional"`
	Out      string `query:"optional"`
}

func (mod *Module) OpRemoveRemoteEndpointLocalPort(ctx *astral.Context, q *ops.Query, args opRemoveRemoteEndpointLocalPort) (err error) {
	endpoint, err := kcp.ParseEndpoint(string(args.Endpoint))
	if err != nil {
		return q.RejectWithCode(4)
	}

	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	err = mod.RemoveEndpointLocalSocket(*endpoint)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return ch.Send(&astral.Ack{})
}
