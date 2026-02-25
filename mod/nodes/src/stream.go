package nodes

import (
	"errors"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes/src/frames"
	"github.com/cryptopunkscc/astrald/sig"
)

type Stream struct {
	*frames.Stream
	id          astral.Nonce
	createdAt   time.Time
	conn        astral.Conn
	pings       sig.Map[astral.Nonce, *Ping]
	checks      atomic.Int32
	outbound    bool
	pingTimeout time.Duration
	wakeCh      chan struct{}
}

type Ping struct {
	sentAt time.Time
	pong   chan struct{}
}

func newStream(conn astral.Conn, id astral.Nonce, outbound bool) *Stream {
	link := &Stream{
		id:          id,
		conn:        conn,
		createdAt:   time.Now(),
		Stream:      frames.NewStream(conn),
		outbound:    outbound,
		pingTimeout: defaultPingTimeout,
		wakeCh:      make(chan struct{}, 1),
	}

	go link.pingLoop()

	return link
}

func (s *Stream) LocalIdentity() *astral.Identity {
	return s.conn.LocalIdentity()
}

func (s *Stream) RemoteIdentity() *astral.Identity {
	return s.conn.RemoteIdentity()
}

func (s *Stream) CloseWithError(err error) error {
	if err != nil {
		return s.Stream.CloseWithError(err)
	}

	return s.Stream.CloseWithError(errors.New("link closed"))
}

func (s *Stream) Network() string {
	if c, ok := s.conn.(exonet.Conn); ok {
		if e := c.RemoteEndpoint(); e != nil {
			return e.Network()
		}
		if e := c.LocalEndpoint(); e != nil {
			return e.Network()
		}
	}
	return "unknown"
}

func (s *Stream) LocalEndpoint() exonet.Endpoint {
	if c, ok := s.conn.(exonet.Conn); ok {
		return c.LocalEndpoint()
	}
	return nil
}

func (s *Stream) RemoteEndpoint() exonet.Endpoint {
	if c, ok := s.conn.(exonet.Conn); ok {
		return c.RemoteEndpoint()
	}

	return nil
}

func (s *Stream) Write(frame frames.Frame) (err error) {
	if _, ok := frame.(*frames.Ping); !ok {
		s.check()
	}
	return s.Stream.Write(frame)
}

func (s *Stream) Ping() (time.Duration, error) {
	var nonce = astral.NewNonce()

	p, ok := s.pings.Set(nonce, &Ping{
		sentAt: time.Now(),
		pong:   make(chan struct{}),
	})
	if !ok {
		return -1, errors.New("duplicate sessionId")
	}
	defer s.pings.Delete(nonce)

	err := s.Write(&frames.Ping{
		Nonce: nonce,
	})
	if err != nil {
		return -1, err
	}
	p.sentAt = time.Now()

	select {
	case <-p.pong:
		return time.Since(p.sentAt), nil
	case <-time.After(s.pingTimeout):
		return -1, errors.New("ping timeout")
	case <-s.Stream.Done():
		return -1, s.Err()
	}
}

func (s *Stream) pong(nonce astral.Nonce) (time.Duration, error) {
	p, ok := s.pings.Delete(nonce)
	if !ok {
		return -1, errors.New("invalid sessionId")
	}
	d := time.Since(p.sentAt)
	close(p.pong)
	return d, nil
}

// Wake triggers a ping on the next loop iteration.
func (s *Stream) Wake() {
	select {
	case s.wakeCh <- struct{}{}:
	default:
	}
}

func (s *Stream) check() {
	if s.checks.Swap(2) != 0 {
		return
	}

	select {
	case s.wakeCh <- struct{}{}:
	default:
	}
}

func (s *Stream) pingLoop() {
	select {
	case <-s.Stream.Done():
		return
	case <-time.After(time.Duration(rand.Int63n(int64(pingJitter)))):
	}

	for {
		if s.checks.Load() > 0 {
			time.Sleep(activeInterval)
		} else {
			select {
			case <-s.Stream.Done():
				return
			case <-s.wakeCh:
			}
		}

		if s.Err() != nil {
			return
		}

		if _, err := s.Ping(); err != nil {
			s.CloseWithError(err)
			return
		}

		if s.checks.Load() > 0 {
			s.checks.Add(-1)
		}
	}
}
