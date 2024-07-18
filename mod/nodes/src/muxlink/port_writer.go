package muxlink

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mux"
	"github.com/cryptopunkscc/astrald/astral"
	"sync"
	"time"
)

var _ astral.SecureWriteCloser = &PortWriter{}

const defaultMaxFrameSize = 1024 * 8
const debugBufferUnderruns = false

type PortWriter struct {
	*astral.SourceField
	sync.Mutex
	link         *Link
	port         int
	err          error
	maxFrameSize int
}

func NewPortWriter(link *Link, port int) *PortWriter {
	return &PortWriter{
		link:         link,
		port:         port,
		maxFrameSize: defaultMaxFrameSize,
		SourceField:  astral.NewSourceField(nil),
	}
}

func (w *PortWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return
	}

	defer w.link.Touch()

	w.Lock()
	defer w.Unlock()

	if w.err != nil {
		return 0, w.err
	}

	w.link.health.Check()

	for {
		var frameLen = len(p)
		if frameLen > w.maxFrameSize {
			frameLen = w.maxFrameSize
		}

		var waitCh = make(chan struct{})
		go func() {
			t0 := time.Now()
			for {
				select {
				case <-time.After(100 * time.Millisecond):
					if debugBufferUnderruns {
						fmt.Printf("BUFFER UNDERRUN: %s port %d: %v bytes waiting for %s\n",
							w.link.RemoteIdentity().Fingerprint(),
							w.port,
							frameLen,
							time.Since(t0).Round(10*time.Millisecond),
						)
					}

				case <-waitCh:
					return
				}
			}
		}()

		err = w.link.remoteBuffers.wait(w.port, frameLen)
		close(waitCh)
		if err != nil {
			return 0, err
		}

		if err = w.link.write(w.port, p[:frameLen]); err != nil {
			return n, err
		}

		n += frameLen
		p = p[frameLen:]
		if len(p) == 0 {
			return n, nil
		}
	}
}

func (w *PortWriter) Identity() id.Identity {
	return w.link.RemoteIdentity()
}

func (w *PortWriter) MaxFrameSize() int {
	return w.maxFrameSize
}

func (w *PortWriter) SetMaxFrameSize(maxFrameSize int) {
	w.maxFrameSize = maxFrameSize
}

func (w *PortWriter) Close() error {
	w.Lock()
	defer w.Unlock()

	if w.err != nil {
		return nil
	}

	w.err = ErrPortClosed

	return w.link.mux.Write(mux.Frame{Port: w.port})
}

func (w *PortWriter) Transport() any {
	return w.link.Transport()
}

func (w *PortWriter) Link() *Link {
	return w.link
}

func (w *PortWriter) Port() int {
	return w.port
}

func (w *PortWriter) BufferSize() int {
	size, _ := w.link.remoteBuffers.size(w.port)
	return size
}
