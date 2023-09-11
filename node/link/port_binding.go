package link

import (
	"github.com/cryptopunkscc/astrald/mux"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/streams"
	"sync/atomic"
)

type PortBinding struct {
	*net.OutputField
	async *streams.AsyncWriter
	link  *CoreLink
	port  atomic.Int32
}

func NewPortBinding(output net.SecureWriteCloser, link *CoreLink) *PortBinding {
	binding := &PortBinding{
		link:  link,
		async: streams.NewAsyncWriter(output, portBufferSize),
	}
	binding.OutputField = net.NewOutputField(binding, output)

	binding.async.SetAfterFlush(func(bytes []byte) {
		if p := binding.port.Load(); p != 0 {
			link.control.GrowBuffer(int(p), len(bytes))
		}
	})

	return binding
}

func (binding *PortBinding) HandleMux(event mux.Event) {
	switch event := event.(type) {
	case mux.Bind:
		binding.port.Store(int32(event.Port))

	case mux.Frame:
		binding.handleFrame(event)

	case mux.Unbind:
		binding.link.control.Reset(int(binding.port.Swap(0)))
		binding.async.Close()
	}
}

func (binding *PortBinding) Link() *CoreLink {
	return binding.link
}

func (binding *PortBinding) Transport() net.SecureConn {
	return binding.link.Transport()
}

func (binding *PortBinding) Port() int {
	return int(binding.port.Load())
}

func (binding *PortBinding) Used() int {
	return binding.async.Used()
}

func (binding *PortBinding) BufferSize() int {
	return binding.async.BufferSize()
}

func (binding *PortBinding) SetOutput(output net.SecureWriteCloser) error {
	binding.async.SetWriter(output)
	return binding.OutputField.SetOutput(output)
}

func (binding *PortBinding) handleFrame(frame mux.Frame) {
	// register link activity
	binding.link.Touch()

	// check EOF
	if frame.IsEmpty() {
		frame.Mux.Unbind(frame.Port)
		return
	}

	// add chunk to the buffer
	if _, err := binding.async.Write(frame.Data); err != nil {
		binding.link.CloseWithError(err)
	}
}
