package nodes

import (
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
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
	// todo: probably state now can be int32 change it when holding stateCond
	stateCond *sync.Cond   // broadcasts on every state transition
	state     atomic.Int32 // connection state

	stream *Stream       // stream the connection is attached to
	bytes  atomic.Uint64 // total bytes transferred (read + write)

	reader  io.Reader      //  io.Reader
	writer  io.WriteCloser // io.WriteCloser
	onClose func()
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
		stateCond: sync.NewCond(&sync.Mutex{}),
	}
}

func (c *session) Read(p []byte) (int, error) {
	c.stateCond.L.Lock()
	for {
		switch c.state.Load() {
		case stateOpen, stateMigrating:
			c.stateCond.L.Unlock()
			return c.reader.Read(p)
		case stateRouting:
			c.stateCond.Wait()
		default:
			c.stateCond.L.Unlock()
			return 0, errors.New("session closed")
		}
	}
}

func (c *session) Write(p []byte) (int, error) {
	c.stateCond.L.Lock()
	for {
		switch c.state.Load() {
		case stateOpen:
			c.stateCond.L.Unlock()
			return c.writer.Write(p)
		case stateRouting, stateMigrating:
			c.stateCond.Wait()
		default:
			c.stateCond.L.Unlock()
			return 0, errors.New("session closed")
		}
	}
}

func (c *session) Open(s *Stream, reader io.Reader, writer io.WriteCloser, onClose func()) {
	c.stateCond.L.Lock()
	defer c.stateCond.L.Unlock()

	c.stream = s
	c.reader = reader
	c.writer = writer
	c.onClose = onClose
	c.swapState(stateRouting, stateOpen)
}

func (c *session) Close() error {
	if !c.swapState(stateOpen, stateClosed) && !c.swapState(stateMigrating, stateClosed) {
		return nodes.ErrInvalidSessionState
	}

	if c.state.Load() == stateClosed {
		return nil
	}

	if c.onClose != nil {
		c.onClose()
	}

	if c.writer != nil {
		c.writer.Close()
	}

	return nil
}

func (c *session) swapState(old, new int) bool {
	if !c.state.CompareAndSwap(int32(old), int32(new)) {
		return false
	}
	c.stateCond.Broadcast()
	return true
}

func (s *session) IsOpen() bool {
	return s.state.Load() == stateOpen
}

// note: this method might change its place
func (s *session) CanAutoMigrate() bool {
	return time.Since(s.createdAt) >= minSessionAge || s.bytes.Load() >= minSessionBytes
}

func (c *session) isOnStream(s *Stream) bool {
	c.stateCond.L.Lock()
	defer c.stateCond.L.Unlock()
	return c.stream == s
}
