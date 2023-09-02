package mux

type Event any

type HandlerFunc func(Event)

// Frame represents a single multiplexer frame
type Frame struct {
	Mux  *FrameMux
	Port int
	Data []byte
}

type Bind struct {
	Mux  *FrameMux
	Port int
}

type Unbind struct {
	Mux  *FrameMux
	Port int
}

// IsEmpty returns true if the frame's data length is zero.
func (frame Frame) IsEmpty() bool {
	return len(frame.Data) == 0
}
