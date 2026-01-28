package nat

import (
	"errors"
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/nat"
)

func (client *Client) PairTake(ctx *astral.Context, pair astral.Nonce, onLockOk func() error) error {
	ch, err := client.queryCh(ctx.IncludeZone(astral.ZoneNetwork), nat.MethodPairTake, query.Args{
		"pair":     pair,
		"initiate": false,
	})
	if err != nil {
		return err
	}
	defer ch.Close()

	exchange := nat.NewPairTakeExchange(ch, pair)

	// Lock exchange
	sig, err := exchange.SendReceive(ctx, nat.PairTakeSignalLock)
	if err != nil {
		return err
	}
	switch sig.Signal {
	case nat.PairTakeSignalTypeLockOk:
		if onLockOk != nil {
			if err := onLockOk(); err != nil {
				return err
			}
		}
	case nat.PairTakeSignalTypeLockBusy:
		return nat.ErrPairBusy
	default:
		return fmt.Errorf("unexpected signal in Lock exchange: %s", sig.Signal)
	}

	// Take exchange
	sig, err = exchange.SendReceive(ctx, nat.PairTakeSignalTypeTake)
	if err != nil {
		return err
	}
	switch sig.Signal {
	case nat.PairTakeSignalTypeTakeOk:
		return nil
	case nat.PairTakeSignalTypeTakeErr:
		return errors.New("responder failed to exchange")
	default:
		return fmt.Errorf("unexpected signal in Take exchange: %s", sig.Signal)
	}
}
