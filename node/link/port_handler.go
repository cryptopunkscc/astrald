package link

import (
	"github.com/cryptopunkscc/astrald/mux"
	"io"
)

type PortBinding struct {
	io.WriteCloser
	link *CoreLink
	port int
}

func (h *PortBinding) HandleFrame(frame mux.Frame) {
	defer h.link.Touch()

	if frame.IsEmpty() {
		h.Close()
		h.link.mux.Unbind(h.port)
		h.link.control.Reset(h.port)
		return
	}

	// TODO: buffer incoming data to avoid blocking
	n, _ := h.Write(frame.Data)

	h.link.control.GrowBuffer(h.port, n)
}
