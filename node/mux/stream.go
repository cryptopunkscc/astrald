package mux

// Stream represents a local receiving stream. All incoming frames will be sent to the go channel returned
// by Frames(). Call Close() to free the channel.
type Stream struct {
	id     StreamID
	frames chan Frame
	mux    *Mux
}

// ID returns the ID of the local stream.
func (c *Stream) ID() StreamID {
	return c.id
}

// Frames returns a channel to which incoming frames will be sent.
func (c *Stream) Frames() <-chan Frame {
	return c.frames
}

// Close frees the channel in the multiplexer.
func (c *Stream) Close() error {
	return c.mux.closeChannel(c.id)
}
