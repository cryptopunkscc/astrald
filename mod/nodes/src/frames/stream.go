// Update stream reader/writer to use FrameBlueprints for (de)serialization.
package frames

import (
	"fmt"
	"io"
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/sig"
	"github.com/cryptopunkscc/astrald/streams"
)

const inChanSize = 32

type Stream struct {
	conn io.ReadWriteCloser
	err  sig.Value[error]
	done chan struct{}
	read chan Frame
	mu   sync.Mutex
}

func NewStream(conn io.ReadWriteCloser) *Stream {
	link := &Stream{
		conn: streams.ReadWriteCloseSplit{
			Reader: conn,
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
	_, _ = s.err.Swap(nil, err)
	_ = s.conn.Close()
	return nil
}

func (s *Stream) Read() <-chan Frame {
	return s.read
}

func (s *Stream) Write(frame Frame) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// FrameBlueprints.Write expects an astral.Object; assert the Frame to that interface
	obj, ok := frame.(astral.Object)
	if !ok {
		return fmt.Errorf("frame does not implement astral.Object")
	}

	// Use FrameBlueprints to write the type header + payload
	_, err = FrameBlueprints.Write(s.conn, obj)
	if err != nil {
		_, _ = s.err.Swap(nil, err)
		_ = s.conn.Close()
		return
	}
	return
}

func (s *Stream) reader() {
	var err error
	defer func() {
		_, _ = s.err.Swap(nil, err)
		_ = s.conn.Close()
		close(s.read)
		close(s.done)
	}()

	for {
		obj, _, derr := FrameBlueprints.Read(s.conn)
		if derr != nil {
			err = derr
			return
		}

		frame, ok := obj.(Frame)
		if !ok {
			err = fmt.Errorf("decoded object is not a Frame")
			return
		}

		s.read <- frame
	}
}

var FrameBlueprints = astral.DefaultBlueprints.Indexed([]string{
	"nodes.frames.ping",
	"nodes.frames.query",
	"nodes.frames.read",
	"nodes.frames.response",
	"nodes.frames.data",
	"nodes.frames.migrate",
	"nodes.frames.reset",
})

func init() {
	_ = FrameBlueprints.Add(
		&Ping{},
		&Query{},
		&Response{},
		&Read{},
		&Data{},
		&Migrate{},
		&Reset{},
	)
}
