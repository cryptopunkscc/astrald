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
	}
}

func (b *PortBinding) pushChunk(p []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(p) > b.available() {
		return ErrPortBufferOverflow
	}

	b.chunks = append(b.chunks, p)
	b.used += len(p)

	b.signal()

	return nil
}

func (b *PortBinding) popChunk() ([]byte, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.chunks) == 0 {
		return nil, ErrPortBufferEmpty
	}

	chunk := b.chunks[0]
	b.chunks = b.chunks[1:]
	b.used -= len(chunk)

	return chunk, nil
}

func (b *PortBinding) chunkCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()

	return len(b.chunks)
}

func (b *PortBinding) AfterUnbind() {
	b.pushChunk([]byte{})
}

func (b *PortBinding) signal() {
	select {
	case b.ch <- struct{}{}:
	default:
	}
}

func (b *PortBinding) flusher() {
	defer func() {
		b.target.Close()
		b.link.control.Reset(b.port)
	}()

	for {
		b.wait()
		var flushed int

		for {
			chunk, err := b.popChunk()
			if err != nil {
				break
			}

			// EOF?
			if len(chunk) == 0 {
				return
			}

			n, err := b.target.Write(chunk)
			if len(chunk) != n {
				b.link.CloseWithError(errors.New("partial write on port"))
				return
			}
			if err != nil {
				b.link.CloseWithError(err)
				return
			}

			flushed += n
			if flushed >= defaultMaxFrameSize {
				break
			}
		}

		if flushed > 0 {
			_ = b.link.control.GrowBuffer(b.port, flushed)
		}
	}
}

func (b *PortBinding) wait() error {
	if b.chunkCount() > 0 {
		return nil
	}
	<-b.ch
	return nil
}

func (b *PortBinding) available() int {
	return b.capacity - b.used
}
