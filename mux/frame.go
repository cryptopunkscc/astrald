package mux

// Frame represents a single multiplexer frame
type Frame struct {
	Port int
	Data []byte
}

type FrameHandler interface {
	HandleFrame(Frame)
}

// IsEmpty returns true if the frame's data length is zero.
func (frame Frame) IsEmpty() bool {
	return len(frame.Data) == 0
}

type AfterUnbind interface {
	AfterUnbind()
}
