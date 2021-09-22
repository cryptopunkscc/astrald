package mux

import (
	"encoding/binary"
	"io"
	"sync"
)

// BasicDemux reads mux frames and splits them into separate streams
type BasicDemux struct {
	reader io.Reader
	readMu sync.Mutex
}

func NewBasicDemux(reader io.Reader) *BasicDemux {
	demux := &BasicDemux{
		reader: reader,
	}

	return demux
}

// ReadFrame reads the next frame and returns stream id, bytes read and an error if one occured
func (dem *BasicDemux) ReadFrame(payload []byte) (int, int, error) {
	dem.readMu.Lock()
	defer dem.readMu.Unlock()

	var header [4]byte

	// ReadFrame the frame header
	n, err := io.ReadFull(dem.reader, header[:])
	if err != nil {
		return 0, n, err
	}

	// Parse the header
	id := int(binary.BigEndian.Uint16(header[0:2]))
	payloadLen := int(binary.BigEndian.Uint16(header[2:4]))

	if payloadLen > len(payload) {
		return id, n, ErrBufferTooSmall
	}

	// ReadFrame frame's payload if any
	if payloadLen > 0 {
		n, err := io.ReadFull(dem.reader, payload[:payloadLen])
		if err != nil {
			return id, n, err
		}
	}

	return id, payloadLen, nil
}
