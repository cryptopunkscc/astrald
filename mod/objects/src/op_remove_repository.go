package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
)

type opRemoveRepositoryArgs struct {
	Name string
	In   string `query:"optional"`
	Out  string `query:"optional"`
}

func (mod *Module) OpRemoveRepository(ctx *astral.Context, q *ops.Query, args opRemoveRepositoryArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	err = mod.RemoveRepository(args.Name)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return ch.Send(&astral.Ack{})
}
