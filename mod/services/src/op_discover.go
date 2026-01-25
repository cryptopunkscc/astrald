package services

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
)

type opServiceDiscoverArgs struct {
	Follow bool `query:"optional"`

	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpDiscover(
	ctx *astral.Context,
	q ops.Query,
	args opServiceDiscoverArgs,
) error {
	ch := q.AcceptChannel(channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	updates, err := mod.DiscoverServices(ctx, q.Caller(), args.Follow)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	for update := range updates {
		if update == nil {
			err = ch.Send(&astral.EOS{})
		} else {
			err = ch.Send(update)
		}
		if err != nil {
			return err
		}
	}

	return ch.Send(&astral.EOS{})
}
