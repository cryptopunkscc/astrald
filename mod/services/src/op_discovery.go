package services

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/services"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opServiceDiscoveryArgs struct {
	In     string `query:"optional"`
	Out    string `query:"optional"`
	Follow bool   `query:"optional"`
}

func (mod *Module) OpDiscovery(
	ctx *astral.Context,
	q shell.Query,
	args opServiceDiscoveryArgs,
) error {
	ch := q.AcceptChannel(channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	opts := services.DiscoverOptions{Snapshot: true, Follow: args.Follow}
	snapshot, updates, err := mod.DiscoverServices(ctx, q.Caller(), opts)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	for _, s := range snapshot {
		err := ch.Send(&s)
		if err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}
	}

	err = ch.Send(&astral.EOS{})
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case update, ok := <-updates:
			if !ok {
				return nil
			}

			serviceChange := update
			err := ch.Send(&serviceChange)
			if err != nil {
				return err
			}
		}
	}
}
