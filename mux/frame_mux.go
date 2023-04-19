package mux

import (
	"context"
	"io"
	"sync"
)

type FrameMux struct {
	mu             sync.Mutex
	mux            *RawMux
	portHandlers   map[int]FrameHandlerFunc
	defaultHandler FrameHandlerFunc
	logID          int
	nextPort       int
}

func NewFrameMux(transport io.ReadWriter, defaultHandler FrameHandlerFunc) *FrameMux {
	return &FrameMux{
		mux:            NewRawMux(transport),
		portHandlers:   make(map[int]FrameHandlerFunc),
		defaultHandler: defaultHandler,
	}
}

// Run runs the FrameMux for the duration of the context
func (mux *FrameMux) Run(ctx context.Context) error {
	var frame Frame
	var err error

	defer mux.unbindAll()

	for {
		frame.Port, frame.Data, err = mux.mux.Read()
		if err != nil {
			return io.EOF
		}

		mux.mu.Lock()
		handler, found := mux.portHandlers[frame.Port]
		mux.mu.Unlock()

		if found {
			if frame.EOF() || handler(frame) != nil {
				mux.Unbind(frame.Port)
			}
		} else if mux.defaultHandler != nil {
			_ = mux.defaultHandler(frame)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
}

// Close sends an EOF frame to the specified remote port
func (mux *FrameMux) Close(remotePort int) error {
	return mux.mux.Write(remotePort, []byte{})
}

// Bind binds a local port to a frame handler. The handler will receive frames received on the specified port.
// The last frame a handler receives will be an EOF after which the port is unbound. If the hander returns
// a non nil-error, the port will be unbound.
func (mux *FrameMux) Bind(localPort int, handler FrameHandlerFunc) error {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	if localPort < 0 || localPort > MaxPorts-1 {
		return ErrInvalidPort
	}

	if _, used := mux.portHandlers[localPort]; used {
		return ErrPortInUse
	}

	mux.portHandlers[localPort] = handler

	return nil
}

// BindAny binds a FrameHandler to any avaiable port.
func (mux *FrameMux) BindAny(handler FrameHandler) (port int, err error) {
	return mux.BindAnyFunc(handler.HandleFrame)
}

// BindAnyFunc binds a FrameHandlerFunc to any available port.
func (mux *FrameMux) BindAnyFunc(handler FrameHandlerFunc) (port int, err error) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	port = -1

	for i := 0; i < MaxPorts; i++ {
		p := (mux.nextPort + i) % MaxPorts
		if _, used := mux.portHandlers[p]; !used {
			port = p
			mux.nextPort = p + 1
			break
		}
	}

	if port == -1 {
		return -1, ErrAllPortsUsed
	}

	mux.portHandlers[port] = handler

	return
}

// Unbind unbinds any handler assigned to the specified port.
func (mux *FrameMux) Unbind(port int) error {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	return mux.unbind(port)
}

// Unbind unbinds any handler assigned to the specified port.
func (mux *FrameMux) unbind(port int) error {
	if handler, found := mux.portHandlers[port]; !found {
		return ErrPortNotInUse
	} else {
		_ = handler(Frame{Port: port, Data: []byte{}})
	}

	delete(mux.portHandlers, port)

	return nil
}

// SetDefaultHandlder sets the catch-all FrameHandlerFunc that will receive all frames received on unbound ports.
// Set to nil to clear.
func (mux *FrameMux) SetDefaultHandlder(defaultHandlder FrameHandlerFunc) {
	mux.defaultHandler = defaultHandlder
}

func (mux *FrameMux) unbindAll() {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	for port := range mux.portHandlers {
		_ = mux.unbind(port)
	}
}
