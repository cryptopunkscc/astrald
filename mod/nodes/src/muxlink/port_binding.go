package muxlink

import (
	"github.com/cryptopunkscc/astrald/mux"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/streams"
	"sync/atomic"
)

type PortBinding struct {
	*astral.OutputField
	async *streams.AsyncWriter
	link  *Link
	port  atomic.Int32
}

func NewPortBinding(output astral.SecureWriteCloser, link *Link) *PortBinding {
	binding := &PortBinding{
		link:  link,
		async: streams.NewAsyncWriter(output, portBufferSize),
	}
	binding.OutputField = astral.NewOutputField(binding, output)

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

func (binding *PortBinding) Link() *Link {
	return binding.link
}

func (binding *PortBinding) Transport() astral.Conn {
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

func (binding *PortBinding) SetOutput(output astral.SecureWriteCloser) error {
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
