package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/nat"
)

func (client *Client) PairTake(ctx *astral.Context, pair astral.Nonce, onLocked func() error) error {
	ch, err := client.queryCh(ctx.IncludeZone(astral.ZoneNetwork), nat.MethodPairTake, query.Args{
		"pair":     pair,
		"initiate": false,
	})
	if err != nil {
		return err
	}
	defer ch.Close()

	// Send lock, receive response
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

	if onLocked != nil {
		if err := onLocked(); err != nil {
			return err
		}
	}

	// Send take, receive response
	if err := ch.Send(&nat.PairTakeSignal{Signal: nat.PairTakeSignalTypeTake, Pair: pair}); err != nil {
		return err
	}

	err = ch.Switch(
		nat.ExpectPairTakeSignal(pair, nat.PairTakeSignalTypeTaken, nat.HandleFailedPairTakeSignal),
		channel.PassErrors,
		channel.WithContext(ctx),
	)
	if err != nil {
		return err
	}

	return nil
}
