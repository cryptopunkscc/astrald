package nodes

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/nodes/src/frames"
	"time"
)

type Stream struct {
	RTT    time.Duration
	conn   astral.Conn
	stream *frames.Stream
}

type Ping struct {
	sentAt time.Time
	stream *Stream
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
