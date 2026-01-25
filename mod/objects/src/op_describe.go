package objects

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
)

type opDescribeArgs struct {
	ID   *astral.ObjectID
	Out  string      `query:"optional"`
	Zone astral.Zone `query:"optional"`
}

func (mod *Module) OpDescribe(ctx *astral.Context, q *ops.Query, args opDescribeArgs) (err error) {
	ctx, cancel := ctx.WithIdentity(q.Caller()).IncludeZone(args.Zone).WithTimeout(time.Minute)
	defer cancel()

	ch := channel.New(q.Accept(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	descriptors, err := mod.Describe(ctx, args.ID)
	if err != nil {
		return
	}

	for descriptor := range descriptors {
		err = ch.Send(descriptor)
		if err != nil {
			return
		}
	}

	return ch.Send(&astral.EOS{})
}
