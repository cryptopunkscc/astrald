package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/nat"
)

func (client *Client) PairTake(ctx *astral.Context, pair astral.Nonce, target *astral.Identity) error {
	args := query.Args{"pair": pair}
	if target != nil {
		args["target"] = target.String()
	}

	ch, err := client.queryCh(ctx, nat.MethodPairTake, args)
	if err != nil {
		return err
	}
	defer ch.Close()

	if target != nil {
		// Initiator: wait for Ack
		return ch.Switch(
			channel.ExpectAck,
			func(msg *astral.ErrorMessage) error { return msg },
			channel.WithContext(ctx),
		)
	}

	// Responder: drive the handshake
	if err := ch.Send(&nat.PairTakeSignal{Signal: nat.PairTakeSignalTypeLock, Pair: pair}); err != nil {
		return err
	}

	err = ch.Switch(
		nat.ExpectPairTakeSignal(pair, nat.PairTakeSignalTypeLocked, nat.HandleFailedPairTakeSignal),
		channel.PassErrors,
		channel.WithContext(ctx),
	)
	if err != nil {
		return err
	}

	if err := ch.Send(&nat.PairTakeSignal{Signal: nat.PairTakeSignalTypeTake, Pair: pair}); err != nil {
		return err
	}

	return ch.Switch(
		nat.ExpectPairTakeSignal(pair, nat.PairTakeSignalTypeTaken, nat.HandleFailedPairTakeSignal),
		channel.PassErrors,
		channel.WithContext(ctx),
	)
}
