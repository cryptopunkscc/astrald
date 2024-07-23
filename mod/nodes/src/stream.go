package nodes

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/nodes/src/frames"
	"github.com/cryptopunkscc/astrald/sig"
	"sync/atomic"
	"time"
)

type Stream struct {
	*frames.Stream
	conn   astral.Conn
	pings  sig.Map[astral.Nonce, *Ping]
	checks atomic.Int32
}

type Ping struct {
	sentAt time.Time
	pong   chan struct{}
}

func NewStream(conn astral.Conn) *Stream {
	link := &Stream{
		conn:   conn,
		Stream: frames.NewStream(conn),
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

func (s *Stream) String() string {
	return "stream"
}

func (s *Stream) Write(frame frames.Frame) (err error) {
	if _, ok := frame.(*frames.Ping); !ok {
		s.check()
	}
	return s.Stream.Write(frame)
}

func (s *Stream) Ping() time.Duration {
	var nonce = astral.NewNonce()

	p, ok := s.pings.Set(nonce, &Ping{
		sentAt: time.Now(),
		pong:   make(chan struct{}),
	})
	if !ok {
		return -1
	}
	defer s.pings.Delete(nonce)

	err := s.Write(&frames.Ping{
		Nonce: nonce,
	})
	if err != nil {
		return -1
	}
	p.sentAt = time.Now()

	select {
	case <-p.pong:
		return time.Since(p.sentAt)
	case <-time.After(pingTimeout):
		return -1
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

			if s.Ping() == -1 {
				s.CloseWithError(errors.New("ping timeout"))
				return
			}

			time.Sleep(1 * time.Second)
			if s.checks.Add(-1) == 0 {
				return
			}
		}
	}()
}
