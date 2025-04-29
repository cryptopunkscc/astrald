package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/object"
	"time"
)

type opDescribeArgs struct {
	ID   *object.ID
	Out  string      `query:"optional"`
	Zone astral.Zone `query:"optional"`
}

func (mod *Module) OpDescribe(ctx *astral.Context, q shell.Query, args opDescribeArgs) (err error) {
	if !mod.Auth.Authorize(q.Caller(), objects.ActionRead, args.ID) {
		return q.Reject()
	}

	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	ctx, cancel := ctx.IncludeZone(args.Zone).WithIdentity(q.Caller()).WithTimeout(time.Minute)
	defer cancel()

	descs, err := mod.Describe(ctx, args.ID, nil)
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
