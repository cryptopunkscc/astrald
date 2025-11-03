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

type handoverRole int32

type handoverState int32

const (
	rolePairTaker handoverRole = iota
	rolePairGiver
)

const (
	HandoverStateInit handoverState = iota
	HandoverStateLockExchange
	HandoverStateTakeExchange
	HandoverStateDone
	HandoverStateFailed
)

const (
	// lockTimeout is defined in pair_pool_entry.go (same package)
	initLockTimeout = 5 * time.Second
	takeTimeout     = 5 * time.Second
)

// PairTaker controls the local handover lifecycle.
type PairTaker struct {
	role  handoverRole
	ch    *astral.Channel
	pair  *pairEntry
	peer  *astral.Identity
	err   atomic.Value // error
	state atomic.Int32 // current state
}

func NewPairTaker(role handoverRole, ch *astral.Channel, pair *pairEntry,
	peer *astral.Identity) *PairTaker {
	f := &PairTaker{
		role: role,
		ch:   ch,
		pair: pair,
		peer: peer,
	}
	f.state.Store(int32(HandoverStateInit))
	return f
}

func (f *PairTaker) Run(ctx context.Context) error {
	state := HandoverStateInit
	var err error

	lockedStarted := false // true iff we successfully called beginLock()

	for state != HandoverStateDone {
		if ctx.Err() != nil {
			return f.failErr(ctx.Err())
		}

		switch state {
		case HandoverStateInit, HandoverStateLockExchange:
			// merge Init into LockExchange for initiator
			if f.role == rolePairTaker && state == HandoverStateInit {
				if !f.pair.beginLock() {
					return f.failf("beginLock failed: pair busy")
				}
				lockedStarted = true

				if err = f.writeSignal(ctx, nat.PairHandoverSignalTypeLock, initLockTimeout); err != nil {
					f.rollbackLock(lockedStarted)
					return f.failErr(err)
				}
				f.state.Store(int32(HandoverStateLockExchange))
			}

			// enter LockExchange (both roles)
			next, e := f.doLockExchange(ctx, &lockedStarted)
			f.state.Store(int32(next))
			if e != nil {
				f.rollbackLock(lockedStarted)
				return f.failErr(e)
			}
			state = next

		case HandoverStateTakeExchange:
			next, e := f.doTakeExchange(ctx)
			f.state.Store(int32(next))
			if e != nil {
				f.rollbackLock(lockedStarted)
				return f.failErr(e)
			}
			state = next

		case HandoverStateFailed:
			f.rollbackLock(lockedStarted)
			f.finish(nil)
			return nil

		default: // includes HandoverStateDone
			f.finish(nil)
			return nil
		}
	}

	f.finish(nil)
	return nil
}

func (f *PairTaker) doLockExchange(ctx context.Context, lockedStarted *bool) (handoverState, error) {
	if f.role == rolePairTaker {
		sig, err := f.readPairSignal(ctx, lockTimeout)
		if err != nil {
			return HandoverStateFailed, err
		}
		switch sig.Signal {
		case nat.PairHandoverSignalTypeLockOk:
			if err := f.pair.waitLocked(ctx); err != nil {
				return HandoverStateFailed, err
			}
			return HandoverStateTakeExchange, nil
		case nat.PairHandoverSignalTypeLockBusy:
			return HandoverStateFailed, errors.New("remote busy during lock")
		default:
			return HandoverStateFailed, fmt.Errorf("unexpected signal in lock exchange: %s", sig.Signal)
		}
	} else if f.role != rolePairGiver {
		return HandoverStateFailed, fmt.Errorf("invalid role: %d", f.role)
	}

	// responder
	sig, err := f.readPairSignal(ctx, lockTimeout)
	if err != nil {
		return HandoverStateFailed, err
	}
	if sig.Signal != nat.PairHandoverSignalTypeLock {
		return HandoverStateFailed, fmt.Errorf("expected %s, got %s", nat.PairHandoverSignalTypeLock, sig.Signal)
	}

	if !*lockedStarted { // beginLock only once
		if !f.pair.beginLock() {
			_ = f.writeSignal(ctx, nat.PairHandoverSignalTypeLockBusy, lockTimeout)
			return HandoverStateFailed, errors.New("local busy: beginLock failed")
		}
		*lockedStarted = true
	}
	if err := f.pair.waitLocked(ctx); err != nil {
		return HandoverStateFailed, err
	}
	if err := f.writeSignal(ctx, nat.PairHandoverSignalTypeLockOk, lockTimeout); err != nil {
		return HandoverStateFailed, err
	}
	return HandoverStateTakeExchange, nil
}

func (f *PairTaker) doTakeExchange(ctx context.Context) (handoverState, error) {
	if f.role == rolePairTaker {
		if err := f.writeSignal(ctx, nat.PairHandoverSignalTypeTake, takeTimeout); err != nil {
			return HandoverStateFailed, err
		}
		sig, err := f.readPairSignal(ctx, takeTimeout)
		if err != nil {
			return HandoverStateFailed, err
		}
		switch sig.Signal {
		case nat.PairHandoverSignalTypeTakeOk:
			f.pair.unlock() // return to pool / resume keepalives
			return HandoverStateDone, nil
		case nat.PairHandoverSignalTypeTakeErr:
			return HandoverStateFailed, errors.New("responder failed to take over")
		default:
			return HandoverStateFailed, fmt.Errorf("expected %s or %s, got %s",
				nat.PairHandoverSignalTypeTakeOk, nat.PairHandoverSignalTypeTakeErr, sig.Signal)
		}
	} else if f.role != rolePairGiver {
		return HandoverStateFailed, fmt.Errorf("invalid role: %d", f.role)
	}

	// responder
	sig, err := f.readPairSignal(ctx, takeTimeout)
	if err != nil {
		return HandoverStateFailed, err
	}
	if sig.Signal != nat.PairHandoverSignalTypeTake {
		return HandoverStateFailed, fmt.Errorf("expected %s, got %s", nat.PairHandoverSignalTypeTake, sig.Signal)
	}

	f.pair.unlock()
	if err := f.writeSignal(ctx, nat.PairHandoverSignalTypeTakeOk, takeTimeout); err != nil {
		return HandoverStateFailed, err
	}
	return HandoverStateDone, nil
}

// --- helpers

// readPairSignal filters by PairID, looping until a matching frame is received or an error occurs.
func (f *PairTaker) readPairSignal(ctx context.Context, timeout time.Duration) (*nat.PairHandoverSignal, error) {
	for {
		sig, err := f.readSignal(ctx, timeout)
		if err != nil {
			return nil, err
		}
		if sig.PairID == f.pair.Nonce {
			return sig, nil
		}
	}
}

// readSignal reads a PairHandoverSignal with a timeout and respects context.
func (f *PairTaker) readSignal(ctx context.Context, timeout time.Duration) (*nat.PairHandoverSignal, error) {
	type result struct {
		sig *nat.PairHandoverSignal
		err error
	}
	resCh := make(chan result, 1)

	go func() {
		obj, err := f.ch.Read()
		if err != nil {
			resCh <- result{nil, err}
			return
		}
		sig, ok := obj.(*nat.PairHandoverSignal)
		if !ok {
			resCh <- result{nil, fmt.Errorf("unexpected object type: %T", obj)}
			return
		}
		resCh <- result{sig, nil}
	}()

	select {
	case r := <-resCh:
		return r.sig, r.err
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(timeout):
		return nil, context.DeadlineExceeded
	}
}

// writeSignal sends a PairHandoverSignal and returns early on timeout or context cancel.
func (f *PairTaker) writeSignal(ctx context.Context, s astral.String8, timeout time.Duration) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- f.ch.Write(&nat.PairHandoverSignal{Signal: s, PairID: f.pair.Nonce})
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

func (f *PairTaker) rollbackLock(started bool) {
	if started {
		f.pair.unlock()
	}
}

func (f *PairTaker) failf(format string, a ...any) error {
	return f.failErr(fmt.Errorf(format, a...))
}
func (f *PairTaker) failErr(err error) error {
	if err != nil {
		f.err.Store(err)
	}
	f.finish(err)
	return err
}

// --- lifecycle ---

func (f *PairTaker) finish(err error) {
	if err != nil {
		f.err.Store(err)
	}
}

func (f *PairTaker) Error() error {
	if v := f.err.Load(); v != nil {
		if e, ok := v.(error); ok {
			return e
		}
	}
	return nil
}

// State returns the current FSM state as int32 (internal enum value).
func (f *PairTaker) State() int32 { return f.state.Load() }
