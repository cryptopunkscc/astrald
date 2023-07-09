package link

import (
	"errors"
	"github.com/cryptopunkscc/astrald/mux"
	"io"
	"sync"
)

type PortBinding struct {
	target   io.WriteCloser
	link     *CoreLink
	port     int
	capacity int
	used     int
	chunks   [][]byte
	err      error
	ch       chan struct{}
	mu       sync.Mutex
}

func NewPortBinding(writeCloser io.WriteCloser, link *CoreLink, port int) *PortBinding {
	binding := &PortBinding{
		target:   writeCloser,
		port:     port,
		link:     link,
		capacity: portBufferSize,
		chunks:   make([][]byte, 0),
		ch:       make(chan struct{}, 1),
	}

	go binding.flusher()

	return binding
}

func (b *PortBinding) HandleFrame(frame mux.Frame) {
	// register link activity
	b.link.Touch()

	// check EOF
	if frame.IsEmpty() {
		b.link.mux.Unbind(b.port)
		return
	}

	// add chunk to the buffer
	if err := b.pushChunk(frame.Data); err != nil {
		b.link.CloseWithError(err)
		b.err = err
	}
}

func (b *PortBinding) pushChunk(p []byte) error {
	defer b.signal()

	b.mu.Lock()
	defer b.mu.Unlock()

	if len(p) > b.available() {
		return ErrPortBufferOverflow
	}

	b.chunks = append(b.chunks, p)
	b.used += len(p)

	return nil
}

func (b *PortBinding) AfterUnbind() {
	b.err = io.EOF
	b.signal()
}

func (b *PortBinding) signal() {
	select {
	case b.ch <- struct{}{}:
	default:
	}
}

func (b *PortBinding) flusher() {
	defer b.target.Close()

	for {
		b.wait()

		if err := b.flush(); err != nil {
			return
		}

		if b.err != nil {
			return
		}
	}
}

func (b *PortBinding) wait() error {
	<-b.ch
	return nil
}

// flush flushes all chunks in the buffer
func (b *PortBinding) flush() error {
	for {
		b.mu.Lock()
		if len(b.chunks) == 0 {
			b.mu.Unlock()
			return nil
		}
		chunk := b.chunks[0]
		b.mu.Unlock()

		// TODO: add timeout
		n, err := b.target.Write(chunk)
		if err != nil {
			return err
		}
		if n != len(chunk) {
			return errors.New("partial write")
		}

		b.mu.Lock()
		b.chunks = b.chunks[1:]
		b.used -= n
		b.mu.Unlock()

		_ = b.link.control.GrowBuffer(b.port, n)
	}
}

func (b *PortBinding) closeWithError(e error) {
	b.err = e
}

func (b *PortBinding) available() int {
	return b.capacity - b.used
}
