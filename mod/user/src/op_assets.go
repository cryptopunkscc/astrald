package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
)

type opAssetsArgs struct {
	Out string `query:"optional"`
}

func (mod *Module) OpAssets(ctx *astral.Context, q *routing.IncomingQuery, args opAssetsArgs) (err error) {
	ch := channel.New(q.AcceptRaw(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	for _, asset := range mod.Assets() {
		err = ch.Send(asset)
		if err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}
	}

	return ch.Send(&astral.EOS{})
}
