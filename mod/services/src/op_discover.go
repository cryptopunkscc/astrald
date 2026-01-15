package services

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opServiceDiscoverArgs struct {
	In     string `query:"optional"`
	Out    string `query:"optional"`
	Follow bool   `query:"optional"`
}

func (mod *Module) OpDiscover(
	ctx *astral.Context,
	q shell.Query,
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
