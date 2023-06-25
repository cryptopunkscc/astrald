package link

import (
	"github.com/cryptopunkscc/astrald/mux"
	"github.com/cryptopunkscc/astrald/net"
	"io"
	"sync"
)

var _ io.WriteCloser = &PortWriter{}
var _ net.Linker = &PortWriter{}

type PortWriter struct {
	sync.Mutex
	link *CoreLink
	port int
	err  error
}

func NewPortWriter(link *CoreLink, port int) *PortWriter {
	return &PortWriter{link: link, port: port}
}

func (w *PortWriter) Write(p []byte) (n int, err error) {
	defer w.link.Touch()

	w.Lock()
	defer w.Unlock()

	if w.err != nil {
		return 0, w.err
	}

	for {
		var frameLen = len(p)
		if frameLen > mux.MaxFrameSize {
			frameLen = mux.MaxFrameSize
		}

		// TODO: stop waiting and close the connection after timeout
		if err = w.link.remoteBuffers.wait(w.port, frameLen); err != nil {
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
