package nodes

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/nodes/frames"
)

const maxPayloadSize = 8192
const defaultBufferSize = 4 * 1024 * 1024

const (
	stateRouting = iota
	stateOpen
	stateMigrating
	stateClosed
)

var _ io.WriteCloser = &session{}

type session struct {
	Nonce          astral.Nonce
	RemoteIdentity *astral.Identity
	relayID        *astral.Identity // non-nil if this session is routed through a relay
	Outbound       bool
	Query          string
	createdAt      time.Time
	res            chan uint8

	state  atomic.Int32  // connection state
	stream *Stream       // stream the connection is attached to; guarded by mu
	bytes  atomic.Uint64 // total bytes transferred (read + write)

	reader *sessionReader
	writer *sessionWriter

	mu                     sync.Mutex
	migratingTo            *Stream
	migratedCh             chan struct{} // closed when migration completes or is cancelled
	migrateFrameSent       atomic.Bool
	migrateFrameReceivedCh chan struct{} // closed when the remote's Migrate frame arrives (responder side)
}

func (c *session) Identity() *astral.Identity {
	return c.RemoteIdentity
}

func newSession(n astral.Nonce) *session {
	if n == 0 {
		n = astral.NewNonce()
	}

	return &session{
		Nonce:     n,
		createdAt: time.Now(),
		res:       make(chan uint8, 1),
	}
}

func (c *session) Read(p []byte) (int, error) {
	return c.reader.Read(p)
}

func (c *session) Write(p []byte) (int, error) {
	return c.writer.Write(p)
}

func (c *session) closeBuffers() {
	if c.reader != nil {
		c.reader.Close()
	}
	if c.writer != nil {
		c.writer.Close()
	}
}

func (c *session) Pause() {
	c.reader.Pause()
	c.writer.Pause()
}

func (c *session) Resume() {
	c.reader.Resume()
	c.writer.Resume()
}

func (c *session) Close() error {
	c.CancelMigration()
	if !c.swapState(stateOpen, stateClosed) {
		return nodes.ErrInvalidSessionState
	}

	c.mu.Lock()
	stream := c.stream
	c.mu.Unlock()

	stream.Write(&frames.Reset{Nonce: c.Nonce})
	c.closeBuffers()

	return nil
}

func (c *session) swapState(old, new int) bool {
	return c.state.CompareAndSwap(int32(old), int32(new))
}

func (s *session) IsOpen() bool {
	return s.state.Load() == stateOpen
}

func (c *session) isOnStream(s *Stream) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.stream == s
}

func (s *session) CanMigrate() bool {
	return s.IsOpen() && (time.Since(s.createdAt) >= minSessionAge || s.bytes.Load() >= minSessionBytes)
}

func (c *session) Migrate(s *Stream) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.stream.RemoteIdentity().IsEqual(s.RemoteIdentity()) {
		return fmt.Errorf("identity mismatch")
	}

	if !c.swapState(stateOpen, stateMigrating) {
		return errors.New("cannot migrate non-open session")
	}

	c.migratingTo = s
	c.migratedCh = make(chan struct{})
	c.migrateFrameReceivedCh = make(chan struct{})
	c.migrateFrameSent.Store(false)

	return nil
}

// writeMigrateFrame sends a migration marker over the current (old) stream.
func (c *session) writeMigrateFrame() error {
	c.mu.Lock()
	stream := c.stream
	c.mu.Unlock()

	if stream == nil {
		return fmt.Errorf("no current stream")
	}

	if err := stream.Write(&frames.Migrate{Nonce: c.Nonce}); err != nil {
		return err
	}
	c.migrateFrameSent.Store(true)
	return nil
}

// CompleteMigration finishes the migration by switching to the target stream and reopening the session.
func (c *session) CompleteMigration() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.migratingTo == nil {
		return errors.New("no target stream")
	}

	c.stream = c.migratingTo
	c.migratingTo = nil

	if !c.swapState(stateMigrating, stateOpen) {
		return nodes.ErrInvalidMigrationState
	}

	if c.migratedCh != nil {
		close(c.migratedCh)
		c.migratedCh = nil
	}

	return nil
}

func (c *session) CancelMigration() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state.Load() != stateMigrating {
		return
	}

	c.migratingTo = nil
	c.migrateFrameSent.Store(false)

	if c.migratedCh != nil {
		close(c.migratedCh)
		c.migratedCh = nil
	}

	c.swapState(stateMigrating, stateOpen)
}

// signalMigrateFrameReceived closes migrateFrameReceivedCh to unblock WaitMigrateFrameReceived.
func (c *session) signalMigrateFrameReceived() {
	c.mu.Lock()
	ch := c.migrateFrameReceivedCh
	c.migrateFrameReceivedCh = nil
	c.mu.Unlock()

	if ch != nil {
		close(ch)
	}
}

// WaitMigrateFrameReceived blocks until the remote's Migrate frame has been received (responder side).
func (c *session) WaitMigrateFrameReceived(ctx context.Context) error {
	c.mu.Lock()
	ch := c.migrateFrameReceivedCh
	c.mu.Unlock()

	if ch == nil {
		return errors.New("not migrating")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ch:
		return nil
	}
}

// WaitOpen blocks until the migration completes or the context is cancelled.
func (c *session) WaitOpen(ctx context.Context) error {
	c.mu.Lock()
	ch := c.migratedCh
	c.mu.Unlock()

	if ch == nil {
		if c.IsOpen() {
			return nil
		}
		return errors.New("not migrating")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ch:
		if c.IsOpen() {
			return nil
		}
		return errors.New("migration cancelled")
	}
}
