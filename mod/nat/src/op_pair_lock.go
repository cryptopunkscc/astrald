package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/nat"
)

type opPairLockArgs struct {
	Pair astral.Nonce

	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpPairLock(ctx *astral.Context, q *ops.Query, args opPairLockArgs) error {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	pair, err := mod.pool.Take(args.Pair)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	if !pair.BeginLock() {
		return ch.Send(astral.Err(nat.ErrPairBusy))
	}

	if err := pair.WaitLocked(ctx); err != nil {
		return ch.Send(astral.Err(err))
	}

	return ch.Send(&astral.Ack{})
}
