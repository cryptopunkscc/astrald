package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
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
	ctx, cancel := ctx.WithIdentity(q.Caller()).IncludeZone(args.Zone).WithTimeout(time.Minute)
	defer cancel()

	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	list, err := mod.Describe(ctx, args.ID, nil)
	if err != nil {
		return
	}

	for so := range list {
		err = ch.Write(so.Object)
		if err != nil {
			return
		}
	}

	return
}
