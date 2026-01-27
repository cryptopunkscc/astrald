package nat

import (
	"errors"
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nat"
)

func (client *Client) PairTake(ctx *astral.Context, pair astral.Nonce, onLockOk func() error) error {
	ch, err := client.PairTakeCh(ctx, pair, false)
	if err != nil {
		return err
	}
	defer ch.Close()

	exchange := nat.NewPairTakeExchange(ch, pair)

	// Lock exchange
	sig, err := exchange.SendReceive(ctx, nat.PairHandoverSignalTypeLock)
	if err != nil {
		return err
	}
	switch sig.Signal {
	case nat.PairHandoverSignalTypeLockOk:
		if onLockOk != nil {
			if err := onLockOk(); err != nil {
				return err
			}
		}
	case nat.PairHandoverSignalTypeLockBusy:
		return nat.ErrPairBusy
	default:
		return fmt.Errorf("unexpected signal in Lock exchange: %s", sig.Signal)
	}

	// Take exchange
	sig, err = exchange.SendReceive(ctx, nat.PairHandoverSignalTypeTake)
	if err != nil {
		return err
	}
	switch sig.Signal {
	case nat.PairHandoverSignalTypeTakeOk:
		return nil
	case nat.PairHandoverSignalTypeTakeErr:
		return errors.New("responder failed to exchange")
	default:
		return fmt.Errorf("unexpected signal in Take exchange: %s", sig.Signal)
	}
}
