package shell

import (
	"slices"
	"strings"

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

	list := mod.root.Spec()
	slices.SortFunc(list, func(a, b ops.OpSpec) int {
		return strings.Compare(a.Name, b.Name)
	})

	for _, o := range list {
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
