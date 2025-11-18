package nat

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nat"
)

type pairTakerRole int32

type takeState int32

const (
	roleTakePairInitiator pairTakerRole = iota
	roleTakePairResponder
)

const (
	TakeStateLockExchange takeState = iota
	TakeStateTakeExchange
	TakeStateDone
	TakeStateFailed
)

const (
	takeTimeout = 5 * time.Second
)

// PairTaker synchronizes port pairs
type PairTaker struct {
	role        pairTakerRole
	ch          *astral.Channel
	pair        *Pair
	err         atomic.Value // error
	state       atomic.Int32 // current state
	lockStarted bool         // tracks if BeginLock was called for rollback
}

func NewPairTaker(role pairTakerRole, ch *astral.Channel, pair *Pair) *PairTaker {
	f := &PairTaker{
		role: role,
		ch:   ch,
		pair: pair,
	}
	return f
}

func (f *PairTaker) Run(ctx context.Context) error {
	f.state.Store(int32(TakeStateLockExchange))
	f.lockStarted = false

	if f.role == roleTakePairInitiator {
		if !f.pair.BeginLock() {
			return f.fail(fmt.Errorf("BeginLock failed: Pair busy"))
		}
		f.lockStarted = true
		if err := f.writeSignal(ctx, nat.PairHandoverSignalTypeLock, f.pair.LockTimeout()); err != nil {
			return f.fail(err)
		}
	}

	if err := f.doLockExchange(ctx); err != nil {
		return f.fail(err)
	}

	f.state.Store(int32(TakeStateTakeExchange))
	if err := f.doTakeExchange(ctx); err != nil {
		return f.fail(err)
	}

	f.state.Store(int32(TakeStateDone))
	return nil
}

func (f *PairTaker) doLockExchange(ctx context.Context) error {
	if f.role == roleTakePairInitiator {
		action := func(sig *nat.PairTakeSignal) error {
			switch sig.Signal {
			case nat.PairHandoverSignalTypeLockOk:
				return f.pair.WaitLocked(ctx)
			case nat.PairHandoverSignalTypeLockBusy:
				return nat.ErrPairBusy
			default:
				return fmt.Errorf("unexpected signal in Lock exchange: %s", sig.Signal)
			}
		}
		return f.exchange(ctx, f.pair.LockTimeout(), nil, nil, action)
	}

	action := func(sig *nat.PairTakeSignal) error {
		if !f.lockStarted {
			if !f.pair.BeginLock() {
				_ = f.writeSignal(ctx, nat.PairHandoverSignalTypeLockBusy, f.pair.LockTimeout())
				return nat.ErrPairBusy
			}
			f.lockStarted = true
		}
		if err := f.pair.WaitLocked(ctx); err != nil {
			return err
		}
		return f.writeSignal(ctx, nat.PairHandoverSignalTypeLockOk, f.pair.LockTimeout())
	}
	expectedLock := astral.String8(nat.PairHandoverSignalTypeLock)
	return f.exchange(ctx, f.pair.LockTimeout(), nil, &expectedLock, action)
}

func (f *PairTaker) doTakeExchange(ctx context.Context) error {
	if f.role == roleTakePairInitiator {
		action := func(sig *nat.PairTakeSignal) error {
			if sig.Signal == nat.PairHandoverSignalTypeTakeErr {
				return errors.New("responder failed to exchange")
			}

			return nil
		}
		sendTake := astral.String8(nat.PairHandoverSignalTypeTake)
		expectTakeOk := astral.String8(nat.PairHandoverSignalTypeTakeOk)
		return f.exchange(ctx, takeTimeout, &sendTake, &expectTakeOk, action)
	}

	action := func(sig *nat.PairTakeSignal) error {
		return f.writeSignal(ctx, nat.PairHandoverSignalTypeTakeOk, takeTimeout)
	}
	expectedTake := astral.String8(nat.PairHandoverSignalTypeTake)
	return f.exchange(ctx, takeTimeout, nil, &expectedTake, action)
}

// readSignal reads PairTakeSignal frames until one for this f.Pair.Nonce arrives,
// honoring context and per-attempt timeout for each read.
func (f *PairTaker) readSignal(ctx context.Context, timeout time.Duration) (*nat.PairTakeSignal, error) {
	for {
		type result struct {
			sig *nat.PairTakeSignal
			err error
		}
		resCh := make(chan result, 1)

		go func() {
			obj, err := f.ch.Read()
			if err != nil {
				resCh <- result{nil, err}
				return
			}
			sig, ok := obj.(*nat.PairTakeSignal)
			if !ok {
				resCh <- result{nil, fmt.Errorf("unexpected object type: %T", obj)}
				return
			}
			resCh <- result{sig, nil}
		}()

		select {
		case r := <-resCh:
			if r.err != nil {
				return nil, r.err
			}

			if r.sig.Pair != f.pair.Nonce {
				return nil, fmt.Errorf("mismatched Pair id  %v (expected %v)",
					r.sig.Pair, f.pair.Nonce)
			}

			return r.sig, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(timeout):
			return nil, context.DeadlineExceeded
		}
	}
}

func (f *PairTaker) exchange(
	ctx context.Context,
	timeout time.Duration,
	send, expect *astral.String8,
	action func(*nat.PairTakeSignal) error,
) error {
	if send != nil {
		if err := f.writeSignal(ctx, *send, timeout); err != nil {
			return err
		}
	}

	sig, err := f.readSignal(ctx, timeout)
	if err != nil {
		return err
	}
	if expect != nil && sig.Signal != *expect {
		return fmt.Errorf("expected %s, got %s", *expect, sig.Signal)
	}
	return action(sig)
}

// writeSignal sends a PairTakeSignal and returns early on timeout or context cancel.
func (f *PairTaker) writeSignal(ctx context.Context, s astral.String8, timeout time.Duration) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- f.ch.Write(&nat.PairTakeSignal{Signal: s, Pair: f.pair.Nonce})
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(timeout):
		return context.DeadlineExceeded
	}
}

// fail handles error recording, state marking, and cleanup in one place.
func (f *PairTaker) fail(err error) error {
	if err == nil {
		return nil
	}
	f.state.Store(int32(TakeStateFailed))
	f.err.Store(err)

	return err
}

func (f *PairTaker) Error() error {
	if v := f.err.Load(); v != nil {
		if e, ok := v.(error); ok {
			return e
		}
	}
	return nil
}
