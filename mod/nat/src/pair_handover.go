package nat

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

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
		if ctx.Err() != nil { // abort early on context cancellation
			f.state.Store(int32(HandoverStateFailed))
			f.fail(ctx.Err())
			return ctx.Err()
		}

		handler, ok := handlers[state]
		if !ok {
			f.state.Store(int32(HandoverStateFailed))
			return errors.New("invalid handover state")
		}

		next, err := handler(ctx)
		// store next state for observability even on error paths
		f.state.Store(int32(next))
		if err != nil {
			f.fail(err)
			return err
		}
		state = next
	}

	f.finish(nil)
	return nil
}

// --- State handlers ---

func (f *PairHandoverFSM) handleInit(ctx context.Context) (handoverState, error) {
	// Initiator: begin local lock, then request peer to lock.
	if f.role == roleInitiator {
		if !f.pair.beginLock() {
			return HandoverStateFailed, errors.New("beginLock failed: pair not idle or busy")
		}
		// Send initial Lock request with PairID.
		lockCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := f.sendCtx(lockCtx, nat.PairHandoverSignalTypeLock); err != nil {
			f.rollbackLock()
			return HandoverStateFailed, err
		}
		return HandoverStateLockExchange, nil
	}

	// Responder: wait for Lock in the next state.
	return HandoverStateLockExchange, nil
}

func (f *PairHandoverFSM) handleLockExchange(ctx context.Context) (handoverState, error) {
	lockExchangeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if f.role == roleInitiator {
		// Wait for responder to report it's locked.
		sig, err := f.recvCtx(lockExchangeCtx)
		if err != nil {
			f.rollbackLock()
			return HandoverStateFailed, err
		}
		switch sig.Signal {
		case nat.PairHandoverSignalTypeLockOk:
			// Remote locked; now wait for our local lock to complete.
			if err := f.pair.waitLocked(ctx); err != nil {
				f.rollbackLock()
				return HandoverStateFailed, err
			}
			return HandoverStateTakeExchange, nil
		case nat.PairHandoverSignalTypeLockBusy:
			f.rollbackLock()
			return HandoverStateFailed, errors.New("remote busy during lock")
		default:
			f.rollbackLock()
			return HandoverStateFailed, fmt.Errorf("unexpected signal in lock exchange: %s", sig.Signal)
		}
	}

	// Responder path.
	sig, err := f.recvCtx(lockExchangeCtx)
	if err != nil {
		return HandoverStateFailed, err
	}
	if sig.Signal != nat.PairHandoverSignalTypeLock {
		return HandoverStateFailed, fmt.Errorf("expected %s, got %s", nat.PairHandoverSignalTypeLock, sig.Signal)
	}

	// Try to enter locking locally.
	if !f.pair.beginLock() {
		_ = f.sendCtx(lockExchangeCtx, nat.PairHandoverSignalTypeLockBusy)
		return HandoverStateFailed, errors.New("local busy: beginLock failed")
	}

	// Wait until locally locked.
	if err := f.pair.waitLocked(ctx); err != nil {
		f.rollbackLock()
		return HandoverStateFailed, err
	}

	// Notify initiator we're locked.
	if err := f.sendCtx(lockExchangeCtx, nat.PairHandoverSignalTypeLockOk); err != nil {
		f.rollbackLock()
		return HandoverStateFailed, err
	}
	return HandoverStateTakeExchange, nil
}

func (f *PairHandoverFSM) handleTakeExchange(ctx context.Context) (handoverState, error) {
	takeExchangeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if f.role == roleInitiator {
		// Ask responder to take over.
		if err := f.sendCtx(takeExchangeCtx, nat.PairHandoverSignalTypeTake); err != nil {
			f.rollbackLock()
			return HandoverStateFailed, err
		}

		sig, err := f.recvCtx(takeExchangeCtx)
		if err != nil {
			f.rollbackLock()
			return HandoverStateFailed, err
		}
		switch sig.Signal {
		case nat.PairHandoverSignalTypeTakeOk:
			// Finalize: unlock local pair to resume normal operation.
			f.pair.unlock()
			return HandoverStateDone, nil
		case nat.PairHandoverSignalTypeTakeErr:
			f.rollbackLock()
			return HandoverStateFailed, errors.New("responder failed to take over")
		default:
			f.rollbackLock()
			return HandoverStateFailed, fmt.Errorf("expected %s or %s, got %s", nat.PairHandoverSignalTypeTakeOk, nat.PairHandoverSignalTypeTakeErr, sig.Signal)
		}
	}

	// Responder path: wait for Take, then unlock and ack.
	sig, err := f.recvCtx(takeExchangeCtx)
	if err != nil {
		f.rollbackLock()
		return HandoverStateFailed, err
	}
	if sig.Signal != nat.PairHandoverSignalTypeTake {
		f.rollbackLock()
		return HandoverStateFailed, fmt.Errorf("expected %s, got %s", nat.PairHandoverSignalTypeTake, sig.Signal)
	}

	// Unlock local pair to resume normal operation, then confirm.
	f.pair.unlock()
	if err := f.sendCtx(takeExchangeCtx, nat.PairHandoverSignalTypeTakeOk); err != nil {
		// Already unlocked, but log or handle
		fmt.Printf("Failed to send TakeOk after unlock: %v\n", err)
		return HandoverStateFailed, err
	}
	return HandoverStateDone, nil
}

func (f *PairHandoverFSM) handleFailed(ctx context.Context) (handoverState, error) {
	// Cleanup: rollback if needed
	f.rollbackLock()
	_ = ctx
	return HandoverStateDone, nil
}

func (f *PairHandoverFSM) handleDone(ctx context.Context) (handoverState, error) {
	_ = ctx
	return HandoverStateDone, nil
}

// --- IO helpers ---

func (f *PairHandoverFSM) recvCtx(ctx context.Context) (*nat.PairHandoverSignal, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		type out struct {
			obj any
			err error
		}
		ch := make(chan out, 1)
		go func() { o, e := f.ch.Read(); ch <- out{o, e} }()
		select {
		case r := <-ch:
			if r.err != nil {
				return nil, r.err
			}
			sig, ok := r.obj.(*nat.PairHandoverSignal)
			if !ok {
				return nil, fmt.Errorf("unexpected type %T", r.obj)
			}
			if sig.PairID != f.pair.Nonce {
				// Ignore non-matching PairID, loop for next
				continue
			}
			return sig, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (f *PairHandoverFSM) sendCtx(ctx context.Context, signal astral.String8) error {
	type out struct{ err error }
	ch := make(chan out, 1)
	go func() { ch <- out{f.ch.Write(&nat.PairHandoverSignal{Signal: signal, PairID: f.pair.Nonce})} }()
	select {
	case r := <-ch:
		return r.err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// --- lifecycle ---

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
	f.closeOnce.Do(func() {
		if f.ch != nil {
			_ = f.ch.Close()
		}
		close(f.done)
	})
}

func (f *PairHandoverFSM) rollbackLock() {
	if f.pair.state.Load() == stateInLocking || f.pair.state.Load() == stateLocked {
		f.pair.unlock() // Revert to idle and resume keepalives
	}
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

func (f *PairHandoverFSM) State() handoverState { return handoverState(f.state.Load()) }
