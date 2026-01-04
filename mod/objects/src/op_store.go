package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opStoreArgs struct {
	Repo string `query:"optional"`
	In   string `query:"optional"`
	Out  string `query:"optional"`
}

func (mod *Module) OpStore(ctx *astral.Context, q shell.Query, args opStoreArgs) error {
	ch := channel.New(
		q.Accept(),
		channel.WithFormats(args.In, args.Out),
		channel.AllowUnparsed(true), // allow unparsed objects
	)
	defer ch.Close()

	return ch.Collect(func(object astral.Object) error {
		objectID, err := mod.Store(ctx, args.Repo, object)
		if err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}

		return ch.Send(objectID)
	})

}
