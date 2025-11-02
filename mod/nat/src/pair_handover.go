package nat

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nat"
)

type handoverRole int32

type handoverState int32

const (
	roleInitiator handoverRole = iota
	roleResponder
)

const (
	HandoverStateInit handoverState = iota
	HandoverStateLockExchange
	HandoverStateTakeExchange
	HandoverStateDone
	HandoverStateFailed
)

type handoverStateFn func(ctx context.Context) (handoverState, error)

// PairHandoverFSM controls the local handover lifecycle.
type PairHandoverFSM struct {
	role handoverRole
	ch   *astral.Channel
	pair *pairEntry
	peer *astral.Identity

	done      chan struct{}
	closeOnce sync.Once
	err       atomic.Value // error
	state     atomic.Int32 // current state
}

func NewPairHandoverFSM(role handoverRole, ch *astral.Channel, pair *pairEntry, peer *astral.Identity) *PairHandoverFSM {
	f := &PairHandoverFSM{
		role: role,
		ch:   ch,
		pair: pair,
		peer: peer,
		done: make(chan struct{}),
	}
	f.state.Store(int32(HandoverStateInit))
	return f
}

// Run drives the handover flow using a state machine similar to traversal.go.
func (f *PairHandoverFSM) Run(ctx context.Context) error {
	handlers := map[handoverState]handoverStateFn{
		HandoverStateInit:         f.handleInit,
		HandoverStateLockExchange: f.handleLockExchange,
		HandoverStateTakeExchange: f.handleTakeExchange,
		HandoverStateFailed:       f.handleFailed,
		HandoverStateDone:         f.handleDone,
	}

	state := HandoverStateInit
	for state != HandoverStateDone {
		handler, ok := handlers[state]
		if !ok {
			return errors.New("invalid handover state")
		}

		next, err := handler(ctx)
		if err != nil {
			f.fail(err)
			return err
		}
		state = next
		f.state.Store(int32(state))
	}

	f.finish(nil)
	return nil
}

// State handlers
func (f *PairHandoverFSM) handleInit(_ context.Context) (handoverState, error) {
	if f.role == roleInitiator {
		// initiator requests both parties to enter locking phase
		err := f.ch.Write(&nat.PairHandoverSignal{Signal: nat.PairHandoverSignalTypeLock, PairID: f.pair.Nonce})
		if err != nil {
			return HandoverStateFailed, err
		}
		return HandoverStateLockExchange, nil
	}

	// responder waits for first signal
	return HandoverStateLockExchange, nil
}

func (f *PairHandoverFSM) handleLockExchange(ctx context.Context) (handoverState, error) {
	if f.role == roleInitiator {
		// Start local locking. If already in use, fail.
		if !f.pair.beginLock() {
			return HandoverStateFailed, errors.New("local lock failed")
		}

		// Wait until local lock completes (socket drained and closed)
		if err := f.pair.waitLocked(ctx); err != nil {
			return HandoverStateFailed, err
		}

		// Announce that we're locked
		if err := f.ch.Write(&nat.PairHandoverSignal{Signal: nat.PairHandoverSignalTypeLockOk, PairID: f.pair.Nonce}); err != nil {
			return HandoverStateFailed, err
		}

		// Wait for remote to acknowledge its lock
		sig, err := f.recv()
		if err != nil {
			return HandoverStateFailed, err
		}
		switch sig.Signal {
		case nat.PairHandoverSignalTypeLockOk:
			return HandoverStateTakeExchange, nil
		case nat.PairHandoverSignalTypeLockBusy:
			return HandoverStateFailed, errors.New("remote busy")
		default:
			return HandoverStateFailed, fmt.Errorf("unexpected %s", sig.Signal)
		}

	} else {
		// Wait for initiator's lock request
		sig, err := f.recv()
		if err != nil {
			return HandoverStateFailed, err
		}
		if sig.Signal != nat.PairHandoverSignalTypeLock {
			return HandoverStateFailed, errors.New("expected lock")
		}

		// Try to enter locking locally
		if !f.pair.beginLock() {
			_ = f.ch.Write(&nat.PairHandoverSignal{Signal: nat.PairHandoverSignalTypeLockBusy, PairID: f.pair.Nonce})
			return HandoverStateFailed, nil
		}

		// Wait until locked locally
		if err := f.pair.waitLocked(ctx); err != nil {
			return HandoverStateFailed, err
		}

		// Notify initiator that we're locked
		if err := f.ch.Write(&nat.PairHandoverSignal{Signal: nat.PairHandoverSignalTypeLockOk, PairID: f.pair.Nonce}); err != nil {
			return HandoverStateFailed, err
		}

		// And wait for initiator's confirmation it is locked
		sig, err = f.recv()
		if err != nil {
			return HandoverStateFailed, err
		}
		if sig.Signal != nat.PairHandoverSignalTypeLockOk {
			return HandoverStateFailed, fmt.Errorf("expected lock_ok, got %v", sig.Signal)
		}
		return HandoverStateTakeExchange, nil
	}
}

func (f *PairHandoverFSM) handleTakeExchange(_ context.Context) (handoverState, error) {
	if f.role == roleInitiator {
		// After both sides are locked, initiator requests the responder to take over
		if err := f.ch.Write(&nat.PairHandoverSignal{Signal: nat.PairHandoverSignalTypeTake, PairID: f.pair.Nonce}); err != nil {
			return HandoverStateFailed, err
		}

		sig, err := f.recv()
		if err != nil {
			return HandoverStateFailed, err
		}
		if sig.Signal != nat.PairHandoverSignalTypeTakeOk {
			return HandoverStateFailed, errors.New("expected take_ok")
		}

		f.pair.expire()
		return HandoverStateDone, nil
	}

	// responder
	sig, err := f.recv()
	if err != nil {
		return HandoverStateFailed, err
	}
	if sig.Signal != nat.PairHandoverSignalTypeTake {
		return HandoverStateFailed, errors.New("expected take")
	}

	f.pair.expire()
	_ = f.ch.Write(&nat.PairHandoverSignal{Signal: nat.PairHandoverSignalTypeTakeOk, PairID: f.pair.Nonce})
	return HandoverStateDone, nil
}

func (f *PairHandoverFSM) handleFailed(_ context.Context) (handoverState, error) {
	return HandoverStateDone, nil
}

func (f *PairHandoverFSM) handleDone(_ context.Context) (handoverState, error) {
	return HandoverStateDone, nil
}

func (f *PairHandoverFSM) recv() (*nat.PairHandoverSignal, error) {
	obj, err := f.ch.Read()
	if err != nil {
		return nil, err
	}
	sig, ok := obj.(*nat.PairHandoverSignal)
	if !ok {
		return nil, fmt.Errorf("unexpected type %T", obj)
	}
	return sig, nil
}

func (f *PairHandoverFSM) fail(err error) {
	if err != nil {
		f.err.Store(err)
	}
	f.finish(err)
}

func (f *PairHandoverFSM) finish(err error) {
	if err != nil {
		f.err.Store(err)
	}
	f.closeOnce.Do(func() { close(f.done) })
}

func (f *PairHandoverFSM) Error() error {
	if v := f.err.Load(); v != nil {
		if e, ok := v.(error); ok {
			return e
		}
	}
	return nil
}

func (f *PairHandoverFSM) Done() <-chan struct{} { return f.done }

func (f *PairHandoverFSM) State() int { return int(f.state.Load()) }
