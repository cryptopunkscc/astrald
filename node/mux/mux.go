package mux

import (
	"encoding/binary"
	"io"
	"sync"
)

// Mux is a simple multiplexer that can exchange frames over the provided transport. It supports up to 65536 concurrent
// streams since it uses a 16-bit int for stream addressing.
type Mux struct {
	transport io.ReadWriter
	readMu    sync.Mutex
	writeMu   sync.Mutex
}

// New instantiates a new multiplexer over the provided transport
func New(transport io.ReadWriter) *Mux {
	return &Mux{transport: transport}
}

// Read reads the next frame and returns stream id, bytes read and an error if one occured
func (mux *Mux) Read(buf []byte) (int, int, error) {
	mux.readMu.Lock()
	defer mux.readMu.Unlock()

	var header [4]byte

	// Read the frame header
	n, err := io.ReadFull(mux.transport, header[:])
	if err != nil {
		return 0, n, err
	}

	// Parse the header
	id := binary.BigEndian.Uint16(header[0:2])
	frameLen := int(binary.BigEndian.Uint16(header[2:4]))

	if frameLen > len(buf) {
		return int(id), n, ErrBufferTooSmall
	}

	// Read frame's payload if any
	if frameLen > 0 {
		n, err := io.ReadFull(mux.transport, buf[0:frameLen])
		if err != nil {
			return int(id), n, err
		}
	}

	return int(id), frameLen, nil
}

// Write writes a frame to the mux
func (mux *Mux) Write(id int, buf []byte) error {
	if (id < 0) || (id >= MaxStreams) {
		return ErrInvalidStreamID
	}

	mux.writeMu.Lock()
	defer mux.writeMu.Unlock()

	if len(buf) > MaxPayload {
		return ErrBufferTooBig
	}

	var err error
	var header [4]byte
	var payloadLen = uint16(len(buf))

	// Construct the header
	binary.BigEndian.PutUint16(header[0:2], uint16(id))
	binary.BigEndian.PutUint16(header[2:4], payloadLen)

	// Write the header
	_, err = mux.transport.Write(header[:])
	if err != nil {
		return err
	}

	// Write the payload
	if payloadLen == 0 {
		return nil
	}
	_, err = mux.transport.Write(buf[0:payloadLen])

	return err
}
