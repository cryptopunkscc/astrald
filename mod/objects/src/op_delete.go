package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opDeleteArgs struct {
	ID   *astral.ObjectID `query:"optional"`
	Repo string           `query:"optional"`
	Out  string           `query:"optional"`
	Zone *astral.Zone     `query:"optional"`
}

func (mod *Module) OpDelete(ctx *astral.Context, q shell.Query, args opDeleteArgs) (err error) {
	// prepare the context
	ctx = ctx.WithIdentity(q.Caller())
	if args.Zone == nil {
		ctx = ctx.WithZone(astral.ZoneAll)
	} else {
		ctx = ctx.WithZone(*args.Zone)
	}

	ch := channel.New(q.Accept(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	// look up the repository
	repo, err := mod.GetRepository(args.Repo)
	if err != nil || repo == nil {
		return q.RejectWithCode(8)
	}

	// if an ID was provided, delete a single object
	if args.ID != nil {
		err := mod.Delete(ctx, args.Repo, args.ID)
		if err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}
		return ch.Send(&astral.Ack{})
	}

	// otherwise read objects ids from the channel
	return ch.Handle(ctx, func(object astral.Object) {
		switch object := object.(type) {
		case *astral.ObjectID:
			err := mod.Delete(ctx, args.Repo, object)
			if err != nil {
				ch.Send(astral.NewError(err.Error()))
			} else {
				ch.Send(&astral.Ack{})
			}

		case *astral.Ack,
			*astral.EOS:

		default: // protocol error
			ch.Send(astral.NewError("protocol error"))
			ch.Close()
		}
	})

}
