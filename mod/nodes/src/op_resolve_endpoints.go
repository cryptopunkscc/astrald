package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opResolveEndpointsArgs struct {
	ID  string
	Out string `query:"optional"`
}

func (mod *Module) OpResolveEndpoints(ctx *astral.Context, q shell.Query, args opResolveEndpointsArgs) (err error) {
	targetID, err := mod.Dir.ResolveIdentity(args.ID)
	if err != nil {
		return q.RejectWithCode(2)
	}

	endpoints, err := mod.ResolveEndpoints(ctx.WithIdentity(q.Caller()), targetID)
	if err != nil {
		mod.log.Error("resolve endpoints: %v", err)
		return q.RejectWithCode(astral.CodeInternalError)
	}

	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	for endpoint := range endpoints {
		err = ch.Write(endpoint)
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}
	}

	return ch.Write(&astral.EOS{})
}
