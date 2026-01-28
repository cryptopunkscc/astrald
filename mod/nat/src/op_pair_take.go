package nat

import (
	"errors"
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
	Pair astral.Nonce

	Initiate bool   `query:"optional"`
	In       string `query:"optional"`
	Out      string `query:"optional"`
}

func (mod *Module) OpPairTake(ctx *astral.Context, q *ops.Query, args opPairTakeArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	pair, err := mod.pool.Take(args.Pair)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	opCtx, cancel := ctx.WithTimeout(pair.LockTimeout() + takeExchangeTimeout)
	defer cancel()

	if args.Initiate {
		remoteIdentity, ok := pair.RemoteIdentity(ctx.Identity())
		if !ok {
			return ch.Send(astral.Err(errors.New("remote endpoint not found")))
		}
		mod.log.Log("taking out pair %v out of pool, starting sync with %v",
			args.Pair, remoteIdentity)

		client := natclient.New(remoteIdentity, astrald.Default())
		if !pair.BeginLock() {
			return ch.Send(astral.Err(nat.ErrPairBusy))
		}

		err = client.PairTake(opCtx, pair.Nonce, func() error {
			return pair.WaitLocked(opCtx)
		})
		if err != nil {
			return ch.Send(astral.Err(err))
		}

		return ch.Send(&pair.TraversedPortPair)
	}

	// Responder flow
	mod.log.Log("taking out pair %v out of pool, starting sync with %v",
		args.Pair, q.Caller())

	pairNonce := pair.Nonce

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
