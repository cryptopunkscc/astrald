package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
)

type opContainsArgs struct {
	Repo string
	ID   *astral.ObjectID `query:"optional"`
	In   string           `query:"optional"`
	Out  string           `query:"optional"`
}

func (mod *Module) OpContains(ctx *astral.Context, q *ops.Query, args opContainsArgs) (err error) {
	ctx = ctx.WithIdentity(q.Caller())

	ch := q.AcceptChannel(channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	repo := mod.GetRepository(args.Repo)
	if repo == nil {
		return ch.Send(astral.NewError("repository not found"))
	}

	if args.ID != nil {
		has, err := repo.Contains(ctx, args.ID)
		if err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}

		return ch.Send((*astral.Bool)(&has))
	}

	return ch.Handle(ctx, func(object astral.Object) {
		switch object := object.(type) {
		case *astral.ObjectID:
			has, err := repo.Contains(ctx, object)
			if err != nil {
				ch.Send(astral.NewError(err.Error()))
			} else {
				ch.Send((*astral.Bool)(&has))
			}

		case *astral.EOS:
			//ignore

		default:
			ch.Send(astral.NewErrUnexpectedObject(object)) // ignore
		}
	})

}
