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
	defer ch.Close()

	caller := q.Caller()

	// One call / one channel: DiscoverServices will emit a snapshot-boundary EOS on its own.
	opts := services.DiscoverOptions{Snapshot: true, Follow: args.Follow}

	stream, err := mod.DiscoverServices(ctx, caller, opts)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	sendObj := func(v astral.Object) error {
		return ch.Send(v)
	}

	if err := serveStream(
		ctx,
		stream,
		sendObj,
		nil, // channel close is enough; DiscoverServices emits EOS for snapshot boundary
	); err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return nil
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
