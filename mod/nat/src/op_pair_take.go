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

type opPairTakeArgs struct {
	Pair   astral.Nonce
	Target string `query:"optional"`

	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpPairTake(ctx *astral.Context, q *ops.Query, args opPairTakeArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	pair, err := mod.pool.Take(args.Pair)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	pairNonce := pair.Nonce

	if args.Target != "" {
		target, err := mod.Dir.ResolveIdentity(args.Target)
		if err != nil {
			return ch.Send(astral.Err(err))
		}

		opCtx, cancel := ctx.WithCancel()
		defer cancel()

		if !pair.BeginLock() {
			return ch.Send(astral.Err(nat.ErrPairBusy))
		}

		natClient := natclient.New(target, astrald.Default())
		err = natClient.PairTake(opCtx, pairNonce, nil)
		if err != nil {
			return ch.Send(astral.Err(err))
		}

		if err := pair.WaitLocked(opCtx); err != nil {
			return ch.Send(astral.Err(err))
		}

		return ch.Send(&astral.Ack{})
	}

	// Responder flow
	opCtx, cancel := ctx.WithTimeout(pair.LockTimeout() + takeExchangeTimeout)
	defer cancel()

	mod.log.Log("taking out pair %v out of pool, starting sync with %v",
		pairNonce, q.Caller())

	// Receive lock
	err = ch.Switch(
		nat.ExpectPairTakeSignal(pairNonce, nat.PairTakeSignalTypeLock, nil),
		channel.PassErrors,
		channel.WithContext(opCtx),
	)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	if !pair.BeginLock() {
		_ = ch.Send(&nat.PairTakeSignal{Signal: nat.PairTakeSignalTypeLocked, Pair: pairNonce, Ok: false, Error: astral.String8(nat.ErrPairBusy.Error())})
		return ch.Send(astral.Err(nat.ErrPairBusy))
	}

	if err := pair.WaitLocked(opCtx); err != nil {
		return ch.Send(astral.Err(err))
	}
	if err := ch.Send(&nat.PairTakeSignal{Signal: nat.PairTakeSignalTypeLocked, Pair: pairNonce, Ok: true}); err != nil {
		return ch.Send(astral.Err(err))
	}

	// Receive take
	err = ch.Switch(
		nat.ExpectPairTakeSignal(pairNonce, nat.PairTakeSignalTypeTake, nil),
		channel.PassErrors,
		channel.WithContext(opCtx),
	)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	if err := ch.Send(&nat.PairTakeSignal{Signal: nat.PairTakeSignalTypeTaken, Pair: pairNonce, Ok: true}); err != nil {
		return ch.Send(astral.Err(err))
	}

	return ch.Send(&pair.TraversedPortPair)
}
