package mux

// Frame represents a single multiplexer frame
type Frame struct {
	Port int
	Data []byte
}

type FrameHandlerFunc func(frame Frame) error

type FrameHandler interface {
	HandleFrame(Frame) error
}

// EOF returns true if the frame represents an EOF.
func (frame Frame) EOF() bool {
	return len(frame.Data) == 0
}
