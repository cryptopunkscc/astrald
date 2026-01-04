package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opLoadArgs struct {
	ID   *astral.ObjectID `query:"optional"`
	Repo string           `query:"optional"`
	Zone *astral.Zone     `query:"optional"`
	Out  string           `query:"optional"`
}

// OpLoad loads an object into memory and writes it to the output. OpLoad verifies the object hash.
func (mod *Module) OpLoad(ctx *astral.Context, q shell.Query, args opLoadArgs) (err error) {
	if args.Zone == nil {
		ctx = ctx.WithZone(astral.ZoneAll)
	} else {
		ctx = ctx.WithZone(*args.Zone)
	}

	ch := channel.New(q.Accept(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	// if an ID was provided, load a single object
	if args.ID != nil {
		o, err := mod.Load(ctx, args.Repo, args.ID)
		if err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}

		return ch.Send(o)
	}

	return ch.Handle(ctx, func(object astral.Object) {
		switch object := object.(type) {
		case *astral.ObjectID:
			o, err := mod.Load(ctx, args.Repo, object)
			if err != nil {
				ch.Send(astral.NewError(err.Error()))
			} else {
				ch.Send(o)
			}

		case *astral.Ack:
			// ignore acks

		case *astral.EOS:
			ch.Close()

		default:
			// protocol error - ignored
			ch.Send(astral.NewError("invalid object id"))
			ch.Close()
		}
	})
}
