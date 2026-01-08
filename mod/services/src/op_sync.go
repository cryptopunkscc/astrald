package services

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opServiceSyncArgs struct {
	ID  *astral.Identity
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpSync(ctx *astral.Context, q shell.Query, args opServiceSyncArgs) (err error) {
	ch := q.AcceptChannel(channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	err = mod.DiscoverRemoteServices(ctx, args.ID, false)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return ch.Send(&astral.EOS{})
}
