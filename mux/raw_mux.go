package mux

import (
	"encoding/binary"
	"github.com/cryptopunkscc/astrald/cslq"
	"io"
	"sync"
)

// MaxPorts - maximum number of ports in the multiplexer (port numbering starts with 0)
const MaxPorts = 1 << 16

// MaxFrameSize - maximum size of a single data frame. Frame length is encoded as uint16,
// so this cannot exceed (1<<16)-1.
const MaxFrameSize = 1<<16 - 1

// frameFormat - cslq format of a data frame
const frameFormat = "s[s]c"

// RawMux reads and writes raw multiplexer frames using provided ReadWriter as transport.
type RawMux struct {
	transport io.ReadWriter
	rmu       sync.Mutex
	wmu       sync.Mutex
	endec     *cslq.Endec
	id        int
}

// NewRawMux returns a new instance of RawMux that uses the provided transport.
func NewRawMux(transport io.ReadWriter) *RawMux {
	return &RawMux{
		transport: transport,
		endec:     cslq.NewEndec(transport),
		id:        int(nextID.Add(1)),
	}
}

// Write writes a single data frame. Port cannot exceed MaxPorts-1. Frame cannot be larger than MaxFrameSize.
// Errors: ErrInvalidPort, ErrFrameTooLarge, ...
func (mux *RawMux) Write(port int, frame []byte) (err error) {
	mux.wmu.Lock()
	defer mux.wmu.Unlock()

	if port < 0 || port > MaxPorts-1 {
		return ErrInvalidPort
	}

	if len(frame) > MaxFrameSize {
		return ErrFrameTooLarge
	}

	err = binary.Write(mux.transport, binary.BigEndian, uint16(port))
	if err != nil {
		return
	}

	err = binary.Write(mux.transport, binary.BigEndian, uint16(len(frame)))
	if err != nil {
		return
	}

	if len(frame) > 0 {
		_, err = mux.transport.Write(frame)
	}

	return err
}

// Read reads a single data frame.
func (mux *RawMux) Read() (port int, frame []byte, err error) {
	mux.rmu.Lock()
	defer mux.rmu.Unlock()

	var p uint16
	var l uint16

	err = binary.Read(mux.transport, binary.BigEndian, &p)
	if err != nil {
		return
	}
	port = int(p)

	err = binary.Read(mux.transport, binary.BigEndian, &l)
	if err != nil {
		return
	}
	if l == 0 {
		return int(p), []byte{}, err
	}

	frame = make([]byte, l)
	_, err = io.ReadFull(mux.transport, frame)

	return
}

// Close closes the transport if it's an io.Closer. Returns ErrCloseUnsupported otherwise.
func (mux *RawMux) Close() error {
	if closer, ok := mux.transport.(io.Closer); ok {
		return closer.Close()
	}

	return ErrCloseUnsupported
}
