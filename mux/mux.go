package mux

import (
	"encoding/binary"
	"io"
	"sync"
)

// Mux is a simple 64k-channel multiplexer
type Mux struct {
	writer io.Writer
	mu     sync.Mutex
}

// NewMux instantiates a new multiplexer that writes to the provided writer
func NewMux(writer io.Writer) *Mux {
	return &Mux{writer: writer}
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

	var err error
	var header [4]byte
	var payloadLen = uint16(len(buf))

	// Construct the header
	binary.BigEndian.PutUint16(header[0:2], uint16(streamID))
	binary.BigEndian.PutUint16(header[2:4], payloadLen)

	// Write the header
	_, err = mux.writer.Write(header[:])
	if err != nil {
		return err
	}

	// Write the payload
	if payloadLen == 0 {
		return nil
	}

	_, err = mux.writer.Write(buf[0:payloadLen])

	return err
}

func (mux *Mux) Stream(streamID int) *OutputStream {
	return NewOutputStream(mux, streamID)
}
