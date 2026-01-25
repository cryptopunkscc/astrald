package services

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
)

type opSyncArgs struct {
	ID     string
	Follow bool   `query:"optional"`
	In     string `query:"optional"`
	Out    string `query:"optional"`
}

func (mod *Module) OpSync(ctx *astral.Context, q *ops.Query, args opSyncArgs) (err error) {
	ch := q.AcceptChannel(channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	// resolve the target identity
	targetID, err := mod.Dir.ResolveIdentity(args.ID)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	ctx, cancel := ctx.WithZone(astral.ZoneNetwork).WithCancel()

	// cancel the sync on any channel activity
	go func() {
		_, _ = ch.Receive()
		cancel()
	}()

	// run the updater
	err = mod.syncServices(ctx, targetID, args.Follow)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return ch.Send(&astral.Ack{})
}
