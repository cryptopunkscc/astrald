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
	mod  *Module
	sess *session
	role migrateRole
	ch   *astral.Channel

	local *astral.Identity
	peer  *astral.Identity

	sessionId astral.Nonce
	streamId  astral.Nonce
}

func (m *sessionMigrator) Run(ctx *astral.Context) error {
	m.mod.log.Log("session_migrator Run start role %v session %v stream %v", m.role, m.sessionId, m.streamId)
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
	m.mod.log.Log("session_migrator Run end session %v stream %v", m.sessionId, m.streamId)
	return nil
}

// handleMigrating:
// - Initiator: send migrate_begin, go to WaitingAck
// - Responder: recv migrate_begin, send migrate_ready, go to WaitingMarker
func (m *sessionMigrator) handleMigrating(ctx *astral.Context) (migrateState, error) {
	if m.role == RoleInitiator {
		m.mod.log.Log("session_migrator initiator sending BEGIN %v %v", m.sessionId, m.streamId)
		if err := m.ch.Write(&nodes.SessionMigrateSignal{Signal: nodes.
			MigrateSignalTypeBegin, Nonce: m.sessionId}); err != nil {
			return StateFailed, err
		}
		return StateWaitingAck, nil
	}

	m.mod.log.Log("session_migrator responder waiting BEGIN %v %v", m.sessionId, m.streamId)
	sig, err := m.recv()
	if err != nil {
		return StateFailed, err
	}
	if err = m.verify(sig, nodes.MigrateSignalTypeBegin); err != nil {
		return StateFailed, err
	}

	// Resolve local target stream and enter migrating
	if m.streamId == 0 {
		return StateFailed, fmt.Errorf("missing streamId on responder")
	}
	localTarget := m.mod.findStreamByID(m.streamId)
	if localTarget == nil {
		return StateFailed, fmt.Errorf("target stream not found")
	}
	m.mod.log.Log("session_migrator responder migrating to stream %v %v", m.sessionId, m.streamId)
	if err := m.sess.Migrate(localTarget); err != nil {
		return StateFailed, err
	}

	// Send ready
	m.mod.log.Log("session_migrator responder sending READY %v %v", m.sessionId, m.streamId)
	if err = m.ch.Write(&nodes.SessionMigrateSignal{Signal: nodes.MigrateSignalTypeReady, Nonce: m.sessionId}); err != nil {
		return StateFailed, err
	}
	return StateWaitingMarker, nil
}

// handleWaitingAck:
// - Initiator: wait for migrate_ready, then go to WaitingMarker
// - Responder: not used; fall through to WaitingMarker if reached
func (m *sessionMigrator) handleWaitingAck(ctx *astral.Context) (migrateState, error) {
	if m.role == RoleInitiator {
		m.mod.log.Log("session_migrator initiator waiting READY %v %v", m.sessionId, m.streamId)
		sig, err := m.recv()
		if err != nil {
			return StateFailed, err
		}
		if err := m.verify(sig, nodes.MigrateSignalTypeReady); err != nil {
			return StateFailed, err
		}

		// Resolve our local target stream and enter migrating
		if m.streamId == 0 {
			return StateFailed, fmt.Errorf("missing streamId on initiator")
		}
		localTarget := m.mod.findStreamByID(m.streamId)
		if localTarget == nil {
			return StateFailed, fmt.Errorf("target stream not found")
		}
		m.mod.log.Log("session_migrator initiator migrating to stream %v %v", m.sessionId, m.streamId)
		if err := m.sess.Migrate(localTarget); err != nil {
			return StateFailed, err
		}

		// Send migration marker on old stream
		m.mod.log.Log("session_migrator initiator sending MARKER %v %v", m.sessionId, m.streamId)
		if err := m.sess.writeMigrateFrame(); err != nil {
			return StateFailed, err
		}
		return StateWaitingMarker, nil
	}
	return StateWaitingMarker, nil
}

// handleWaitingMarker:
// - Initiator: send migrate_completed (Phase 0 close-out), go to Completed
// - Responder: wait for migrate_completed, then go to Completed
func (m *sessionMigrator) handleWaitingMarker(ctx *astral.Context) (migrateState, error) {
	_ = ctx
	if m.role == RoleInitiator {
		m.mod.log.Log("session_migrator initiator waiting COMPLETED %v %v", m.sessionId, m.streamId)
		sig, err := m.recv()
		if err != nil {
			return StateFailed, err
		}
		if err = m.verify(sig, nodes.MigrateSignalTypeCompleted); err != nil {
			return StateFailed, err
		}
		m.mod.log.Log("session_migrator initiator completing migration %v %v", m.sessionId, m.streamId)
		if err := m.sess.CompleteMigration(); err != nil {
			return StateFailed, err
		}
		m.mod.log.Log("session_migrator initiator completed %v %v", m.sessionId, m.streamId)
		return StateCompleted, nil
	}

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	m.mod.log.Log("session_migrator responder waiting marker application %v %v", m.sessionId, m.streamId)

	for {
		select {
		case <-ctx.Done():
			return StateFailed, ctx.Err()
		case <-ticker.C:
			if m.sess != nil && m.sess.state.Load() == int32(stateOpen) {
				m.mod.log.Log("session_migrator responder sending COMPLETED %v %v", m.sessionId, m.streamId)
				if err := m.ch.Write(&nodes.SessionMigrateSignal{Signal: nodes.MigrateSignalTypeCompleted, Nonce: m.sessionId}); err != nil {
					return StateFailed, err
				}
				m.mod.log.Log("session_migrator responder completed %v %v", m.sessionId, m.streamId)
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
	m.mod.log.Log("session_migrator recv %v %v", sig.Signal, sig.Nonce)
	return sig, nil
}

func (m *sessionMigrator) verify(sig *nodes.SessionMigrateSignal, expected string) error {
	if sig == nil || string(sig.Signal) != expected {
		m.mod.log.Log("session_migrator verify failed expected %v got %v", expected, sig)
		return fmt.Errorf("invalid %v", expected)
	}
	if sig.Nonce != m.sessionId {
		m.mod.log.Log("session_migrator verify nonce mismatch %v %v", sig.Nonce, m.sessionId)
		return fmt.Errorf("sessionId mismatch in %v", expected)
	}
	m.mod.log.Log("session_migrator verify ok %v %v", expected, m.sessionId)
	return nil
}
