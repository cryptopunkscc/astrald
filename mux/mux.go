package mux

import (
	"github.com/cryptopunkscc/astrald/cslq"
	"io"
	"sync"
)

// Mux is a simple 64k-channel multiplexer
type Mux struct {
	*cslq.Encoder
	writer io.Writer
	mu     sync.Mutex
}

// NewMux instantiates a new multiplexer that writes to the provided writer
func NewMux(writer io.Writer) *Mux {
	return &Mux{
		writer:  writer,
		Encoder: cslq.NewEncoder(writer),
	}
}

// Write writes a frame to the mux
func (mux *Mux) Write(streamID int, buf []byte) error {
	if (streamID < 0) || (streamID >= MaxStreams) {
		return ErrInvalidStreamID
	}
	if len(buf) > MaxPayload {
		return ErrBufferTooBig
	}

	mux.mu.Lock()
	defer mux.mu.Unlock()

	return mux.Encode("s[s]c", streamID, buf)
}

func (mux *Mux) Stream(streamID int) *OutputStream {
	return NewOutputStream(mux, streamID)
}
