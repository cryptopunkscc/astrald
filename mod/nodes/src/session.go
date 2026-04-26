package nodes

import (
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
)

const maxPayloadSize = 8192
const defaultBufferSize = 4 * 1024 * 1024

const minSessionAge = 30 * time.Second
const minSessionBytes = 1 * 1024 * 1024 // 1 MB

const (
	stateRouting = iota
	stateOpen
	stateMigrating
	stateClosed
)

var _ io.WriteCloser = &session{}
var _ io.ReadCloser = &session{}

type session struct {
	Nonce          astral.Nonce
	RemoteIdentity *astral.Identity
	relayID        *astral.Identity // non-nil if this session is routed through a relay
	Outbound       bool
	Query          string
	createdAt      time.Time
	res            chan uint8
	cond           *sync.Cond // guards paused, closed, state, stream
	paused         bool
	closed         bool
	state          int32 // purely informational

	stream *Stream       // stream the connection is attached to
	bytes  atomic.Uint64 // total bytes transferred (read + write)

	reader  io.ReadCloser  // io.ReadCloser
	writer  io.WriteCloser // io.WriteCloser
	remove  func()         // removes session from the sessions map
	onClose func()         // sends Reset frame to peer
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
		cond:      sync.NewCond(&sync.Mutex{}),
		paused:    true,
		state:     stateRouting,
	}
}

func (c *session) Read(p []byte) (int, error) {
	c.cond.L.Lock()
	for c.paused && !c.closed {
		c.cond.Wait()
	}
	c.cond.L.Unlock()

	if c.reader == nil {
		return 0, io.EOF
	}
	return c.reader.Read(p)
}

func (c *session) Write(p []byte) (int, error) {
	c.cond.L.Lock()
	for c.paused && !c.closed {
		c.cond.Wait()
	}
	closed := c.closed
	c.cond.L.Unlock()

	if closed {
		return 0, errors.New("session closed")
	}
	return c.writer.Write(p)
}

func (c *session) Open(s *Stream, reader io.ReadCloser, writer io.WriteCloser, onClose func()) {
	c.cond.L.Lock()
	defer c.cond.L.Unlock()

	c.stream = s
	c.reader = reader
	c.writer = writer
	c.onClose = onClose
	c.state = stateOpen
	c.paused = false
	c.cond.Broadcast()
}

func (c *session) Close() error {
	return c.closeWith(c.onClose)
}

// PeerClose closes the session in response to a Reset frame from the peer.
// Unlike Close, it does not fire onClose, so no Reset is sent back.
func (c *session) PeerClose() error {
	return c.closeWith(nil)
}

func (c *session) closeWith(onClose func()) error {
	c.cond.L.Lock()
	if c.closed {
		c.cond.L.Unlock()
		return nil
	}
	remove := c.remove
	c.state = stateClosed
	c.closed = true
	c.cond.Broadcast()
	c.cond.L.Unlock()

	if remove != nil {
		remove()
	}
	if onClose != nil {
		onClose()
	}
	if c.writer != nil {
		c.writer.Close()
	}
	if c.reader != nil {
		c.reader.Close()
	}
	return nil
}

func (c *session) setState(s int32) {
	c.cond.L.Lock()
	defer c.cond.L.Unlock()
	c.state = s
}

func (c *session) getState() int32 {
	c.cond.L.Lock()
	defer c.cond.L.Unlock()
	return c.state
}

func (s *session) IsOpen() bool {
	return s.getState() == stateOpen
}

// note: this method might change its place
func (s *session) CanAutoMigrate() bool {
	return time.Since(s.createdAt) >= minSessionAge || s.bytes.Load() >= minSessionBytes
}

func (c *session) isOnStream(s *Stream) bool {
	c.cond.L.Lock()
	defer c.cond.L.Unlock()
	return c.stream == s
}
