package services

import (
	"context"

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

	snapshotOpts := services.DiscoverOptions{Snapshot: true, Follow: false}
	followOpts := services.DiscoverOptions{Snapshot: false, Follow: args.Follow}

	snapshot, err := mod.DiscoverServices(ctx, caller, snapshotOpts)
	if err != nil {
		return err
	}

	sendEOS := func() error {
		return ch.Send(&astral.EOS{})
	}

	sendServiceChange := func(v services.ServiceChange) error {
		return ch.Send(&v)
	}

	if err := serveStream(
		ctx,
		snapshot,
		sendServiceChange,
		sendEOS,
	); err != nil {
		return err
	}

	if !args.Follow {
		return nil
	}

	stream, err := mod.DiscoverServices(ctx, caller, followOpts)
	if err != nil {
		return err
	}

	return serveStream(
		ctx,
		stream,
		sendServiceChange,
		sendEOS,
	)
}

func serveStream[T any](
	ctx context.Context,
	in <-chan T,
	send func(T) error,
	onClose func() error,
) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case v, ok := <-in:
			if !ok {
				if onClose != nil {
					return onClose()
				}
				return nil
			}
			if err := send(v); err != nil {
				return err
			}
		}
	}
}
