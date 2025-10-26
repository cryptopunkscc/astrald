package nodes

import (
	"fmt"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
)

type migrateRole int

type migrateState int

type stateFn func(*astral.Context) (migrateState, error)

const (
	RoleInitiator migrateRole = iota
	RoleResponder
)

// Unified signaling FSM states
const (
	StateMigrating migrateState = iota
	StateWaitingAck
	StateWaitingMarker
	StateCompleted
	StateFailed
)

type sessionMigrator struct {
	mod   *Module
	sess  *session
	role  migrateRole
	ch    *astral.Channel
	local *astral.Identity
	peer  *astral.Identity
	nonce astral.Nonce
	link  nodes.LinkSelector
}

func (m *sessionMigrator) Run(ctx *astral.Context) error {
	handlers := map[migrateState]stateFn{
		StateMigrating:     m.handleMigrating,
		StateWaitingAck:    m.handleWaitingAck,
		StateWaitingMarker: m.handleWaitingMarker,
	}

	state := StateMigrating
	for state != StateCompleted && state != StateFailed {
		h, ok := handlers[state]
		if !ok {
			return fmt.Errorf("invalid state")
		}
		next, err := h(ctx)
		if err != nil {
			return err
		}
		state = next
	}
	return nil
}

// handleMigrating:
// - Initiator: send migrate_begin, go to WaitingAck
// - Responder: recv migrate_begin, send migrate_ready, go to WaitingMarker
func (m *sessionMigrator) handleMigrating(ctx *astral.Context) (migrateState, error) {
	_ = ctx // reserved for future timeouts
	if m.role == RoleInitiator {
		if err := m.ch.Write(&nodes.SessionMigrateSignal{Signal: nodes.MigrateSignalTypeBegin, Nonce: m.nonce, Link: m.link}); err != nil {
			return StateFailed, err
		}
		return StateWaitingAck, nil
	}

	sig, err := m.recv()
	if err != nil {
		return StateFailed, err
	}
	if err = m.verify(sig, nodes.MigrateSignalTypeBegin); err != nil {
		return StateFailed, err
	}

	// Pick local target stream and migrate
	var localTarget *Stream
	if m.mod != nil && m.sess != nil {
		localTarget = m.mod.peers.pickAltStream(m.sess.stream)
	}
	if localTarget == nil {
		return StateFailed, fmt.Errorf("no local target stream available")
	}

	if err := m.sess.Migrate(localTarget); err != nil {
		return StateFailed, err
	}

	// Send ready with responder local stream id
	if err = m.ch.Write(&nodes.SessionMigrateSignal{Signal: nodes.MigrateSignalTypeReady, Nonce: m.nonce, Link: nodes.LinkSelector{Identity: m.peer, StreamId: astral.Int64(localTarget.id)}}); err != nil {
		return StateFailed, err
	}
	return StateWaitingMarker, nil
}

// handleWaitingAck:
// - Initiator: wait for migrate_ready, then go to WaitingMarker
// - Responder: not used; fall through to WaitingMarker if reached
func (m *sessionMigrator) handleWaitingAck(ctx *astral.Context) (migrateState, error) {
	_ = ctx // reserved for future timeouts
	if m.role == RoleInitiator {
		sig, err := m.recv()
		if err != nil {
			return StateFailed, err
		}
		if err := m.verify(sig, nodes.MigrateSignalTypeReady); err != nil {
			return StateFailed, err
		}
		// Send migration marker on old stream
		if err := m.sess.writeMigrateFrame(); err != nil {
			return StateFailed, err
		}
		return StateWaitingMarker, nil
	}
	// Responder shouldn't typically be here in Phase 0; proceed to marker wait.
	return StateWaitingMarker, nil
}

// handleWaitingMarker:
// - Initiator: send migrate_completed (Phase 0 close-out), go to Completed
// - Responder: wait for migrate_completed, then go to Completed
func (m *sessionMigrator) handleWaitingMarker(ctx *astral.Context) (migrateState, error) {
	_ = ctx // reserved for future timeouts
	if m.role == RoleInitiator {
		sig, err := m.recv()
		if err != nil {
			return StateFailed, err
		}
		if err = m.verify(sig, nodes.MigrateSignalTypeCompleted); err != nil {
			return StateFailed, err
		}
		if err := m.sess.CompleteMigration(); err != nil {
			return StateFailed, err
		}
		return StateCompleted, nil
	}

	// Responder: wait until session returns to open (marker applied via peers.handleMigrate), then send completed
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return StateFailed, ctx.Err()
		case <-ticker.C:
			if m.sess != nil && m.sess.state.Load() == int32(stateOpen) {
				if err := m.ch.Write(&nodes.SessionMigrateSignal{Signal: nodes.MigrateSignalTypeCompleted, Nonce: m.nonce}); err != nil {
					return StateFailed, err
				}
				return StateCompleted, nil
			}
		}
	}
}

func (m *sessionMigrator) recv() (*nodes.SessionMigrateSignal, error) {
	obj, err := m.ch.Read()
	if err != nil {
		return nil, err
	}
	sig, ok := obj.(*nodes.SessionMigrateSignal)
	if !ok {
		return nil, fmt.Errorf("unexpected object type: %T", obj)
	}
	return sig, nil
}

func (m *sessionMigrator) verify(sig *nodes.SessionMigrateSignal, expected string) error {
	if sig == nil || string(sig.Signal) != expected {
		return fmt.Errorf("invalid %s signal", expected)
	}
	if sig.Nonce != m.nonce {
		return fmt.Errorf("nonce mismatch in %s", expected)
	}
	return nil
}
