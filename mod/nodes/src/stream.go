package nodes

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/nodes/src/frames"
	"github.com/cryptopunkscc/astrald/sig"
	"time"
)

type Stream struct {
	conn   astral.Conn
	stream *frames.Stream
	pings  sig.Map[astral.Nonce, *Ping]
}

type Ping struct {
	sentAt time.Time
	pong   chan struct{}
}

func NewStream(conn astral.Conn) *Stream {
	link := &Stream{
		conn:   conn,
		stream: frames.NewStream(conn),
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
		return s.stream.CloseWithError(err)
	}

	return s.stream.CloseWithError(errors.New("link closed"))
}

func (s *Stream) Read() <-chan frames.Frame {
	return s.stream.Read()
}

func (s *Stream) Write(frame frames.Frame) (err error) {
	return s.stream.Write(frame)
}

func (s *Stream) String() string {
	return "stream"
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
