package kcp

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/kcp"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opAddEndpointLocalSocket struct {
	Endpoint  astral.String
	LocalPort astral.Uint16
	Replace   astral.Bool

	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpAddEndpointPortMapping(ctx *astral.Context, q shell.Query, args opAddEndpointLocalSocket) (err error) {
	endpoint, err := kcp.ParseEndpoint(string(args.Endpoint))
	if err != nil {
		return q.RejectWithCode(4)
	}

	ch := astral.NewChannelFmt(q.Accept(), args.In, args.Out)
	defer ch.Close()

	err = mod.SetEndpointLocalSocket(*endpoint, args.LocalPort, args.Replace)
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	return ch.Write(&astral.Ack{})
}
