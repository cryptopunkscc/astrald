package nodes

import (
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
	cond           *sync.Cond // guards paused, closed, stream
	paused         bool
	closed         bool
	state          atomic.Int32

	stream *Stream       // stream the connection is attached to
	bytes  atomic.Uint64 // total bytes transferred (read + write)

	reader io.ReadCloser
	writer io.WriteCloser
	remove func() // removes session from the sessions map
}

func (s *session) Identity() *astral.Identity {
	return s.RemoteIdentity
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
	}
}

func (s *session) Read(p []byte) (int, error) {
	s.cond.L.Lock()
	for s.paused && !s.closed {
		s.cond.Wait()
	}
	s.cond.L.Unlock()

	if s.reader == nil {
		return 0, io.EOF
	}
	n, err := s.reader.Read(p)
	s.bytes.Add(uint64(n))
	return n, err
}

func (s *session) Write(p []byte) (int, error) {
	s.cond.L.Lock()
	for s.paused && !s.closed {
		s.cond.Wait()
	}
	closed := s.closed
	s.cond.L.Unlock()

	if closed {
		return 0, nodes.ErrSessionClosed
	}
	n, err := s.writer.Write(p)
	s.bytes.Add(uint64(n))
	return n, err
}

func (s *session) Setup(stream *Stream, reader io.ReadCloser, writer io.WriteCloser) error {
	s.cond.L.Lock()
	defer s.cond.L.Unlock()

	if s.closed {
		return nodes.ErrSessionClosed
	}

	s.stream = stream
	s.reader = reader
	s.writer = writer
	s.state.Store(stateOpen)
	return nil
}

func (s *session) Open() {
	s.cond.L.Lock()
	defer s.cond.L.Unlock()
	s.paused = false
	s.cond.Broadcast()
}

func (s *session) Close() error {
	s.cond.L.Lock()
	if s.closed {
		s.cond.L.Unlock()
		return nil
	}
	remove := s.remove
	s.state.Store(stateClosed)
	s.closed = true
	s.cond.Broadcast()
	s.cond.L.Unlock()

	if remove != nil {
		remove()
	}
	if s.writer != nil {
		s.writer.Close()
	}
	if s.reader != nil {
		s.reader.Close()
	}
	return nil
}

func (s *session) setState(state int32) {
	s.state.Store(state)
}

func (s *session) getState() int32 {
	return s.state.Load()
}

func (s *session) swapState(old, new int32) bool {
	return s.state.CompareAndSwap(old, new)
}

func (s *session) IsOpen() bool {
	return s.state.Load() == stateOpen
}

// note: this method might change its place
func (s *session) CanAutoMigrate() bool {
	return time.Since(s.createdAt) >= minSessionAge || s.bytes.Load() >= minSessionBytes
}

func (s *session) isOnStream(stream *Stream) bool {
	s.cond.L.Lock()
	defer s.cond.L.Unlock()
	return s.stream == stream
}
