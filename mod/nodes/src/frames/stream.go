package frames

import (
	"bufio"
	"io"
	"sync"

	"github.com/cryptopunkscc/astrald/sig"
	"github.com/cryptopunkscc/astrald/streams"
)

const inChanSize = 32

type Stream struct {
	conn   io.ReadWriteCloser
	peeker *bufio.Reader

	err  sig.Value[error]
	done chan struct{}
	read chan Frame
	mu   sync.Mutex
}

func NewStream(conn io.ReadWriteCloser) *Stream {
	peeker := bufio.NewReader(conn)

	link := &Stream{
		peeker: peeker,
		conn: streams.ReadWriteCloseSplit{
			Reader: peeker,
			Writer: conn,
			Closer: conn,
		},
		done: make(chan struct{}),
		read: make(chan Frame, inChanSize),
	}

	go link.reader()

	return link
}

func (s *Stream) Err() error {
	return s.err.Get()
}

func (s *Stream) CloseWithError(err error) error {
	s.err.Swap(nil, err)
	s.conn.Close()
	return nil
}

func (s *Stream) Read() <-chan Frame {
	return s.read
}

func (s *Stream) Write(frame Frame) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err = frame.WriteTo(s.conn)
	if err != nil {
		s.err.Swap(nil, err)
		s.conn.Close()
		return
	}
	return
}

func (s *Stream) reader() {
	var err error
	defer func() {
		s.err.Swap(nil, err)
		s.conn.Close()
		close(s.read)
		close(s.done)
	}()

	for {
		var opCode []byte

		opCode, err = s.peeker.Peek(1)
		if err != nil {
			return
		}

		var frame Frame

		switch opCode[0] {
		case opPing:
			frame = &Ping{}
		case opQuery:
			frame = &Query{}
		case opResponse:
			frame = &Response{}
		case opRead:
			frame = &Read{}
		case opData:
			frame = &Data{}
		case opMigrate:
			frame = &Migrate{}
		case opReset:
			frame = &Reset{}
		default:
			err = ErrInvalidOpcode
			return
		}

		_, err = frame.ReadFrom(s.conn)
		if err != nil {
			return
		}

		s.read <- frame
	}
}
