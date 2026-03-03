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
	"github.com/cryptopunkscc/astrald/mod/nodes/src/frames"
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
	Outbound       bool
	Query          string
	createdAt      time.Time
	res            chan uint8

	state  atomic.Int32 // connection state
	stream *Stream      // stream the connection is attached to

	rcond *sync.Cond // sync for read functions
	rsize int        // read buffer size
	rused int        // used read buffer
	rbuf  [][]byte   // read buffer

	wcond            *sync.Cond // sync for write functions
	wsize            int        // remote buffer left
	migratingTo      *Stream
	migratedCh       chan struct{} // closed when migration completes or is cancelled
	migrateFrameSent atomic.Bool
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
		wcond:     sync.NewCond(&sync.Mutex{}),
		rcond:     sync.NewCond(&sync.Mutex{}),
		rsize:     defaultBufferSize,
	}
}

func (c *session) growRemoteBuffer(s int) {
	c.wcond.L.Lock()
	defer c.wcond.L.Unlock()

	c.wsize += s
	c.wcond.Broadcast()
}

func (c *session) Write(p []byte) (n int, err error) {
	c.wcond.L.Lock()
	defer c.wcond.L.Unlock()

	for {
		if len(p) == 0 {
			return
		}

		switch c.state.Load() {
		case stateOpen:
		case stateRouting, stateMigrating:
			c.wcond.Wait()
			continue
		default:
			err = errors.New("invalid state")
			return
		}

		if c.wsize == 0 {
			c.wcond.Wait()
			continue
		}

		var l = min(c.wsize, len(p), maxPayloadSize)

		c.wcond.L.Unlock()
		err = c.stream.Write(&frames.Data{
			Nonce:   c.Nonce,
			Payload: p[0:l],
		})
		c.wcond.L.Lock()
		if err != nil {
			return
		}

		c.wsize -= l
		n = n + l
		p = p[l:]
	}
}

func (c *session) pushRead(b []byte) error {
	c.rcond.L.Lock()
	defer c.rcond.L.Unlock()

	if c.rused+len(b) > c.rsize {
		return errors.New("buffer overflow")
	}

	c.rbuf = append(c.rbuf, b)
	c.rused += len(b)

	c.rcond.Broadcast()

	return nil
}

func (c *session) Read(p []byte) (n int, err error) {
	c.rcond.L.Lock()
	defer c.rcond.L.Unlock()

	for {
		if len(p) == 0 {
			return
		}

		if len(c.rbuf) == 0 {
			switch c.state.Load() {
			case stateOpen, stateMigrating:
				c.rcond.Wait()
				continue
			case stateClosed:
				err = errors.New("connection closed")
			default:
				err = errors.New("invalid state")
			}
			return
		}

		b := c.rbuf[0]
		n = min(len(p), len(b))
		copy(p, b[:n])
		b = b[n:]
		c.rused -= n
		if len(b) > 0 {
			c.rbuf[0] = b
		} else {
			c.rbuf = c.rbuf[1:]
		}

		if c.state.Load() == stateOpen {
			c.stream.Write(&frames.Read{
				Nonce: c.Nonce,
				Len:   uint32(n),
			})
		}

		return
	}
}

func (c *session) Close() error {
	if !c.swapState(stateOpen, stateClosed) {
		return errors.New("invalid state")
	}

	c.stream.Write(&frames.Reset{Nonce: c.Nonce})

	return nil
}

func (c *session) swapState(old, new int) bool {
	if !c.state.CompareAndSwap(int32(old), int32(new)) {
		return false
	}
	c.rcond.Broadcast()
	c.wcond.Broadcast()
	return true
}

func (s *session) IsOpen() bool {
	return s.state.Load() == stateOpen
}

func (c *session) Migrate(s *Stream) error {
	c.wcond.L.Lock()
	defer c.wcond.L.Unlock()

	if c.state.Load() != stateOpen {
		return fmt.Errorf(`cannot migrate non-open session`)
	}

	if !c.stream.RemoteIdentity().IsEqual(s.RemoteIdentity()) {
		return fmt.Errorf("identity mismatch")
	}

	if !c.swapState(stateOpen, stateMigrating) {
		return errors.New("cannot migrate non-open session")
	}

	c.migratingTo = s
	c.migratedCh = make(chan struct{})
	c.migrateFrameSent.Store(false)

	return nil
}

// writeMigrateFrame sends a migration marker over the current (old) stream.
func (c *session) writeMigrateFrame() error {
	if c.stream == nil {
		return fmt.Errorf("no current stream")
	}

	c.migrateFrameSent.Store(true)
	return c.stream.Write(&frames.Migrate{Nonce: c.Nonce})
}

// CompleteMigration finishes the migration by switching to the target stream and reopening the session.
func (c *session) CompleteMigration() error {
	c.wcond.L.Lock()
	defer c.wcond.L.Unlock()

	if c.migratingTo == nil {
		return errors.New("no target stream")
	}

	c.stream = c.migratingTo
	c.migratingTo = nil

	if !c.swapState(stateMigrating, stateOpen) {
		if c.state.Load() != stateOpen {
			return errors.New("invalid migration state")
		}
	}

	if c.migratedCh != nil {
		close(c.migratedCh)
		c.migratedCh = nil
	}

	return nil
}

func (c *session) CancelMigration() {
	c.wcond.L.Lock()
	defer c.wcond.L.Unlock()

	if c.state.Load() != stateMigrating {
		return
	}

	c.migratingTo = nil
	c.migrateFrameSent.Store(false)
	c.swapState(stateMigrating, stateOpen)

	if c.migratedCh != nil {
		close(c.migratedCh)
		c.migratedCh = nil
	}
}

// WaitOpen blocks until the migration completes or the context is cancelled.
func (c *session) WaitOpen(ctx context.Context) error {
	c.wcond.L.Lock()
	ch := c.migratedCh
	c.wcond.L.Unlock()

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
