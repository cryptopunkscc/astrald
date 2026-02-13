package nodes

import (
	"context"
	"fmt"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/nodes"
)

type migrateRole int

type migrateState int

type stateFn func(context.Context) (migrateState, error)

const (
	RoleInitiator migrateRole = iota
	RoleResponder
)

const (
	StateMigrating migrateState = iota
	StateWaitingMarker
	StateCompleted
	StateFailed
)

const (
	// NOTE: not sure of the exact values
	migrationTotalTimeout    = 10 * time.Second
	migrateSignalTimeout     = 5 * time.Second
	markerApplicationTimeout = 5 * time.Second
)

type sessionMigrator struct {
	mod  *Module
	sess *session
	role migrateRole
	ch   *channel.Channel

	local *astral.Identity
	peer  *astral.Identity

	sessionId astral.Nonce
	streamId  astral.Nonce
}

func (m *sessionMigrator) Run(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, migrationTotalTimeout)
	defer cancel()

	handlers := map[migrateState]stateFn{
		StateMigrating:     m.handleMigrating,
		StateWaitingMarker: m.handleWaitingMarker,
	}

	state := StateMigrating
	var runErr error
	for state != StateCompleted && state != StateFailed {
		h, ok := handlers[state]
		if !ok {
			runErr = fmt.Errorf("invalid state")
			state = StateFailed
			break
		}

		next, err := h(ctx)
		if err != nil {
			runErr = err
			state = StateFailed
			break
		}
		state = next
	}
	if state == StateFailed {
		m.mod.log.Log("migration: migration cancelled for session %v", m.sessionId)
		m.sess.CancelMigration()
	}
	return runErr
}

// handleMigrating:
// - Initiator: send migrate_begin, wait for migrate_ready, migrate locally, send marker, go to WaitingMarker
// - Responder: recv migrate_begin, send migrate_ready, go to WaitingMarker
func (m *sessionMigrator) handleMigrating(ctx context.Context) (migrateState,
	error) {
	if m.role == RoleInitiator {
		// Send BEGIN
		err := m.writeSignal(ctx, &nodes.SessionMigrateSignal{Signal: nodes.MigrateSignalTypeBegin, Nonce: m.sessionId}, migrateSignalTimeout)
		if err != nil {
			return m.fail(err)
		}
		// Wait READY
		sig, err := m.readSignal(ctx, migrateSignalTimeout)
		if err != nil {
			return m.fail(err)
		}
		err = m.verify(sig, nodes.MigrateSignalTypeReady)
		if err != nil {
			return m.fail(err)
		}
		// Resolve target stream and enter migrating
		if m.streamId == 0 {
			return m.fail(fmt.Errorf("missing stream ID"))
		}
		targetStream := m.mod.findStreamByID(m.streamId)
		if targetStream == nil {
			return m.fail(fmt.Errorf("target stream not found"))
		}

		if !m.sess.stream.RemoteIdentity().IsEqual(targetStream.RemoteIdentity()) {
			return m.fail(fmt.Errorf("identity mismatch"))
		}

		err = m.sess.Migrate(targetStream)
		if err != nil {
			return m.fail(err)
		}
		// Send migration marker on old stream
		err = m.sess.writeMigrateFrame()
		if err != nil {
			return m.fail(err)
		}
		return StateWaitingMarker, nil
	}

	// Responder branch
	sig, err := m.readSignal(ctx, migrateSignalTimeout)
	if err != nil {
		return m.fail(err)
	}
	err = m.verify(sig, nodes.MigrateSignalTypeBegin)
	if err != nil {
		return m.fail(err)
	}

	// Resolve local target stream and enter migrating
	if m.streamId == 0 {
		return m.fail(fmt.Errorf("missing stream ID"))
	}
	targetStream := m.mod.findStreamByID(m.streamId)
	if targetStream == nil {
		return m.fail(fmt.Errorf("target stream not found"))
	}
	// Cannot change identity that we sent to
	if !m.sess.stream.RemoteIdentity().IsEqual(targetStream.RemoteIdentity()) {
		return m.fail(fmt.Errorf("identity mismatch"))
	}
	err = m.sess.Migrate(targetStream)
	if err != nil {
		return m.fail(err)
	}

	// Send READY
	err = m.writeSignal(ctx, &nodes.SessionMigrateSignal{Signal: nodes.MigrateSignalTypeReady, Nonce: m.sessionId}, migrateSignalTimeout)
	if err != nil {
		return m.fail(err)
	}
	return StateWaitingMarker, nil
}

// handleWaitingMarker:
// - Initiator: wait for migrate_completed from responder and complete migration
// - Responder: wait for session to reopen, then send completed
func (m *sessionMigrator) handleWaitingMarker(ctx context.Context) (migrateState,
	error) {
	if m.role == RoleInitiator {
		sig, err := m.readSignal(ctx, migrateSignalTimeout)
		if err != nil {
			return m.fail(err)
		}

		err = m.verify(sig, nodes.MigrateSignalTypeCompleted)
		if err != nil {
			return m.fail(err)
		}

		err = m.sess.CompleteMigration()
		if err != nil {
			return m.fail(err)
		}

		m.mod.log.Log("migration: session migrated %v %v", m.sessionId, m.streamId)
		return StateCompleted, nil
	}

	// Responder: wait for session reopen, then send completed
	ctx, cancel := context.WithTimeout(ctx, markerApplicationTimeout)
	defer cancel()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return m.fail(ctx.Err())
		case <-ticker.C:
			if m.sess != nil && m.sess.IsOpen() {
				m.mod.log.Log("migration: session migrated %v %v", m.sessionId, m.streamId)
				err := m.writeSignal(ctx, &nodes.SessionMigrateSignal{
					Signal: nodes.MigrateSignalTypeCompleted,
					Nonce:  m.sessionId,
				}, migrateSignalTimeout)
				if err != nil {
					return m.fail(err)
				}
				return StateCompleted, nil
			}
		}
	}
}

// readSignal reads a migrate signal with a timeout and respects context cancel.
func (m *sessionMigrator) readSignal(ctx context.Context,
	timeout time.Duration) (*nodes.SessionMigrateSignal, error) {
	type result struct {
		sig *nodes.SessionMigrateSignal
		err error
	}
	resCh := make(chan result, 1)

	go func() {
		obj, err := m.ch.Receive()
		if err != nil {
			resCh <- result{nil, err}
			return
		}
		sig, ok := obj.(*nodes.SessionMigrateSignal)
		if !ok {
			resCh <- result{nil, astral.NewErrUnexpectedObject(obj)}
			return
		}
		resCh <- result{sig, nil}
	}()

	select {
	case <-ctx.Done():
	case <-time.After(timeout):
	case r := <-resCh:
		return r.sig, r.err
	}

	return nil, context.DeadlineExceeded
}

// writeSignal sends a SessionMigrateSignal and returns early on timeout or context cancel.
func (m *sessionMigrator) writeSignal(ctx context.Context,
	obj *nodes.SessionMigrateSignal, timeout time.Duration) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- m.ch.Send(obj)
	}()

	select {
	case <-ctx.Done():
	case <-time.After(timeout):
	case err := <-errCh:
		return err
	}

	return context.DeadlineExceeded
}

func (m *sessionMigrator) verify(sig *nodes.SessionMigrateSignal, expected string) error {
	if sig == nil || string(sig.Signal) != expected {
		return fmt.Errorf("invalid %v", expected)
	}
	if sig.Nonce != m.sessionId {
		return fmt.Errorf("sessionId mismatch in %v", expected)
	}
	return nil
}

func (m *sessionMigrator) fail(reason error) (migrateState, error) {
	_ = m.writeSignal(context.Background(), &nodes.SessionMigrateSignal{Signal: nodes.MigrateSignalTypeAbort, Nonce: m.sessionId}, 1*time.Second)
	return StateFailed, reason
}
