package link

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/mux"
	"github.com/cryptopunkscc/astrald/net"
	"io"
	"sync"
	"time"
)

var _ io.WriteCloser = &PortWriter{}
var _ net.Linker = &PortWriter{}

const defaultMaxFrameSize = 1024 * 8
const debugBufferUnderruns = false

type PortWriter struct {
	sync.Mutex
	link         *CoreLink
	port         int
	err          error
	maxFrameSize int
}

func NewPortWriter(link *CoreLink, port int) *PortWriter {
	return &PortWriter{
		link:         link,
		port:         port,
		maxFrameSize: defaultMaxFrameSize,
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

func (w *PortWriter) Link() net.Link {
	return w.link
}

func (w *PortWriter) Port() int {
	return w.port
}
