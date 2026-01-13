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
	defer func() { _ = ch.Close() }()

	caller := q.Caller()

	opts := services.DiscoverOptions{Snapshot: true, Follow: args.Follow}
	snapshot, updates, err := mod.DiscoverServices(ctx, caller, opts)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	// Snapshot phase.
	for i := range snapshot {
		v := snapshot[i]
		if err := ch.Send(&v); err != nil {
			return err
		}
	}

	// Snapshot boundary.
	if err := ch.Send(&astral.EOS{}); err != nil {
		return err
	}

	// Update phase.
	for {
		select {
		case <-ctx.Done():
			return nil
		case v, ok := <-updates:
			if !ok {
				return nil
			}
			vv := v
			if err := ch.Send(&vv); err != nil {
				return err
			}
		}
	}
}
