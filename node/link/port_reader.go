package link

import (
	"github.com/cryptopunkscc/astrald/mux"
	"io"
	"sync"
	"time"
)

const DefaultPortWriteTimeout = 1 * time.Millisecond
const DefaultPortBufferSize = 4 * mux.MaxFrameSize

// PortReader is a mux.FrameHandler that implements the io.Reader interface. Incoming data is buffered. If
// incoming data overflows the buffer Reader will close with ErrBufferOverflow error.
type PortReader struct {
	cond         *sync.Cond
	frames       [][]byte
	buffered     int
	bufferSize   int
	err          error
	writeTimeout time.Duration
	errorHandler func(error)
}

// NewPortReader is calls NewPortReaderOpts with default buffer size and timeout.
func NewPortReader() *PortReader {
	return NewPortReaderOpts(DefaultPortBufferSize, DefaultPortWriteTimeout)
}

// NewPortReaderOpts returns a new Reader that can be bound to a mux port. If writeTimeout is 0, HandleFrame will
// instantly fail if the buffer is full. If it's -1, HandleFrame will block indefinitely waiting for the buffer to
// clear.
func NewPortReaderOpts(bufferSize int, writeTimeout time.Duration) *PortReader {
	if bufferSize == 0 {
		bufferSize = DefaultPortBufferSize
	}
	return &PortReader{
		bufferSize:   bufferSize,
		cond:         sync.NewCond(&sync.Mutex{}),
		frames:       make([][]byte, 0, 16),
		writeTimeout: writeTimeout,
	}
}

// Read reads from the buffer. If the buffer is empty, it will wait for HandleFrame to write more data to the buffer.
// When an EOF frame is received, after all data from the buffer is read, subsequent Read calls will return io.EOF.
// If the reader was closed due to buffer overlow the error will be ErrBufferOverflow.
func (p *PortReader) Read(buf []byte) (n int, err error) {
	p.cond.L.Lock()
	defer func() {
		p.cond.Signal()
		p.cond.L.Unlock()
	}()

	for (p.buffered == 0) && (p.err == nil) {
		p.cond.Wait()
	}

	if p.buffered == 0 {
		return 0, p.err
	}

	for n < len(buf) {
		if len(p.frames) == 0 {
			break
		}

		var c = copy(buf[n:], p.frames[0])
		if c < len(p.frames[0]) {
			p.frames[0] = p.frames[0][c:]
		} else {
			p.frames = p.frames[1:]
		}
		p.buffered -= c
		n += c
	}

	return n, nil
}

// HandleFrame handles an incoming mux data frame.
func (p *PortReader) HandleFrame(frame mux.Frame) error {
	p.cond.L.Lock()
	defer func() {
		p.cond.Signal()
		p.cond.L.Unlock()
	}()

	if p.err != nil {
		return p.err
	}

	if frame.EOF() {
		return p.setError(io.EOF)
	}

	if p.buffered+len(frame.Data) > p.bufferSize {
		var done = make(chan struct{})
		defer close(done)

		if p.writeTimeout > 0 {
			go func() {
				select {
				case <-time.After(p.writeTimeout):
					p.cond.Signal()
				case <-done:
				}
			}()
		}
		if p.writeTimeout != 0 {
			p.cond.Wait()
		}
		if p.buffered+len(frame.Data) > p.bufferSize {
			return p.setError(ErrBufferOverflow)
		}
	}

	p.frames = append(p.frames, frame.Data)
	p.buffered += len(frame.Data)

	return nil
}

// SetBufferSize sets the maximum amount of data that can be stored in the buffer.
func (p *PortReader) SetBufferSize(bufferSize int) {
	p.bufferSize = bufferSize
}

// SetWriteTimeout sets how long the buffer writer will wait for the buffer to clear before failing
// with ErrBufferOverflow.
func (p *PortReader) SetWriteTimeout(writeTimeout time.Duration) {
	p.writeTimeout = writeTimeout
}

// SetErrorHandler sets the handler that will be called when the reader closes with an error (such as EOF).
func (p *PortReader) SetErrorHandler(errorHandler func(error)) {
	p.errorHandler = errorHandler
}

// setError if called for the first time sets the error for the Reader. Subsequent calls will not change the error.
// Returns the first error that was set.
func (p *PortReader) setError(err error) error {
	if p.err == nil {
		p.err = err
		if p.errorHandler != nil {
			p.errorHandler(err)
		}
	}
	return p.err
}
