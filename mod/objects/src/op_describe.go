package objects

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/object"
	"time"
)

type opDescribeArgs struct {
	ID     object.ID
	Format string      `query:"optional"`
	Zones  astral.Zone `query:"optional"`
}

func (mod *Module) OpDescribe(ctx *astral.Context, q shell.Query, args opDescribeArgs) (err error) {
	if !mod.Auth.Authorize(q.Caller(), objects.ActionRead, &args.ID) {
		return q.Reject()
	}

	if q.Origin() == "network" {
		args.Zones &= ^astral.ZoneVirtual
	}

	scope := astral.DefaultScope()
	scope.Zone = args.Zones

	ch := astral.NewChannel(q.Accept(), args.Format)
	defer ch.Close()

	opCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	descs, err := mod.Describe(opCtx, args.ID, scope)
	if err != nil {
		return
	}

	for so := range descs {
		if !mod.Auth.Authorize(q.Caller(), objects.ActionReadDescriptor, so.Object) {
			continue
		}

		ch.Write(so.Object)
	}

	return
}
