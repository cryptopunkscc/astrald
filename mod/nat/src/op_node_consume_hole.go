package nat

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/nat"
	natclient "github.com/cryptopunkscc/astrald/mod/nat/client"
)

const takeExchangeTimeout = 5 * time.Second

type opNodeConsumeHoleArgs struct {
	Pair   astral.Nonce
	Target string `query:"optional"`

	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpNodeConsumeHole(ctx *astral.Context, q *ops.Query, args opNodeConsumeHoleArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	hole, err := mod.pool.Take(args.Pair)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	holeNonce := hole.Nonce

	if args.Target != "" {
		target, err := mod.Dir.ResolveIdentity(args.Target)
		if err != nil {
			return ch.Send(astral.Err(err))
		}

		opCtx, cancel := ctx.WithCancel()
		defer cancel()

		if !hole.BeginLock() {
			return ch.Send(astral.Err(nat.ErrHoleBusy))
		}

		natClient := natclient.New(target, astrald.Default())
		err = natClient.NodeConsumeHole(opCtx, holeNonce, nil)
		if err != nil {
			return ch.Send(astral.Err(err))
		}

		if err := hole.WaitLocked(opCtx); err != nil {
			return ch.Send(astral.Err(err))
		}

		return ch.Send(&astral.Ack{})
	}

	// Responder flow
	opCtx, cancel := ctx.WithTimeout(hole.LockTimeout() + takeExchangeTimeout)
	defer cancel()

	mod.log.Log("taking out hole %v out of pool, starting sync with %v",
		holeNonce, q.Caller())

	// Receive lock
	err = ch.Switch(
		nat.ExpectConsumeHoleSignal(holeNonce, nat.ConsumeHoleSignalTypeLock, nil),
		channel.PassErrors,
		channel.WithContext(opCtx),
	)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	if !hole.BeginLock() {
		_ = ch.Send(&nat.ConsumeHoleSignal{Signal: nat.ConsumeHoleSignalTypeLocked, Pair: holeNonce, Ok: false, Error: astral.String8(nat.ErrHoleBusy.Error())})
		return ch.Send(astral.Err(nat.ErrHoleBusy))
	}

	if err := hole.WaitLocked(opCtx); err != nil {
		return ch.Send(astral.Err(err))
	}
	if err := ch.Send(&nat.ConsumeHoleSignal{Signal: nat.ConsumeHoleSignalTypeLocked, Pair: holeNonce, Ok: true}); err != nil {
		return ch.Send(astral.Err(err))
	}

	// Receive take
	err = ch.Switch(
		nat.ExpectConsumeHoleSignal(holeNonce, nat.ConsumeHoleSignalTypeTake, nil),
		channel.PassErrors,
		channel.WithContext(opCtx),
	)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	if err := ch.Send(&nat.ConsumeHoleSignal{Signal: nat.ConsumeHoleSignalTypeTaken, Pair: holeNonce, Ok: true}); err != nil {
		return ch.Send(astral.Err(err))
	}

	return ch.Send(&hole.Hole)
}
