package kcp

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/kcp"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opListEndpointsLocalMappingsArgs struct {
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpListEndpointsLocalMappings(ctx *astral.Context, q shell.Query, args opListEndpointsLocalMappingsArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	mappings := mod.GetEndpointsMappings()
	for k, v := range mappings {
		err = ch.Send(&kcp.EndpointLocalMapping{
			Address: k,
			Port:    v,
		})
		if err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}
	}

	return ch.Send(&astral.EOS{})
}
