package nodes

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes/src/frames"
	"github.com/cryptopunkscc/astrald/sig"
	"sync/atomic"
	"time"
)

var lastStreamID atomic.Int32

type Stream struct {
	*frames.Stream
	id        int
	createdAt time.Time
	conn      astral.Conn
	pings     sig.Map[astral.Nonce, *Ping]
	checks    atomic.Int32
	outbound  bool
}

type Ping struct {
	sentAt time.Time
	pong   chan struct{}
}

func newStream(conn astral.Conn, outbound bool) *Stream {
	link := &Stream{
		id:        int(lastStreamID.Add(1)),
		conn:      conn,
		createdAt: time.Now(),
		Stream:    frames.NewStream(conn),
		outbound:  outbound,
	}

	return link
}

func (s *Stream) LocalIdentity() id.Identity {
	return s.conn.LocalIdentity()
}

func (s *Stream) RemoteIdentity() id.Identity {
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

func (s *Stream) String() string {
	if c, ok := s.conn.(exonet.Conn); ok {
		if e := c.RemoteEndpoint(); e != nil {
			return e.Network() + ":" + e.Address()
		}
		if e := c.LocalEndpoint(); e != nil {
			return e.Network() + ":" + e.Address()
		}
	}
	return "stream"
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
		return -1, errors.New("duplicate nonce")
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
	case <-time.After(pingTimeout):
		return -1, errors.New("ping timeout")
	}
}

func (s *Stream) pong(nonce astral.Nonce) error {
	p, ok := s.pings.Delete(nonce)
	if !ok {
		return errors.New("invalid nonce")
	}
	close(p.pong)
	return nil
}

func (s *Stream) check() {
	if s.checks.Swap(2) != 0 {
		return
	}

	go func() {
		for {
			if s.Err() != nil {
				return
			}

			if _, err := s.Ping(); err != nil {
				s.CloseWithError(err)
				return
			}

			time.Sleep(1 * time.Second)
			if s.checks.Add(-1) == 0 {
				return
			}
		}
	}()
}
