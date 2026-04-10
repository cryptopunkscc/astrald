package shell

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
)

type opArgsArgs struct {
	Op  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpSpec(ctx *astral.Context, q *ops.Query, args opArgsArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	ops := mod.root.Spec()
	for _, o := range ops {
		if len(args.Op) > 0 && o.Name != args.Op {
			continue
		}
		err = ch.Send(&o)
		if err != nil {
			return
		}
	}

	return nil
}
