package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opResolveEndpointsArgs struct {
	ID     *astral.Identity
	Format string `query:"optional"`
}

func (mod *Module) OpResolveEndpoints(ctx *astral.Context, q shell.Query, args opResolveEndpointsArgs) (err error) {
	ch := astral.NewChannel(q.Accept(), args.Format)
	defer ch.Close()

	endpoints, err := mod.ResolveEndpoints(ctx.WithIdentity(q.Caller()), args.ID)
	if err != nil {
		mod.log.Error("Failed to resolve endpoints: %v", err)
		return ch.Write(astral.NewError("internal error"))
	}

	for endpoint := range endpoints {
		err = ch.Write(endpoint)
		if err != nil {
			return
		}
	}

	return
}
