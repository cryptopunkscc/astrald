package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
)

type opSetEnabledArgs struct {
	Arg bool   `query:"optional"`
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpSetEnabled(ctx *astral.Context, q *ops.Query, args opSetEnabledArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	mod.SetEnabled(args.Arg)

	return ch.Send(&astral.Ack{})
}
