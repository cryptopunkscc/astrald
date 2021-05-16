package mux

import (
	"encoding/binary"
	"io"
)

// Frame represents a single multiplexer frame
type Frame struct {
	StreamID StreamID // Local StreamID for inbound frames, remote StreamID for outbound frames
	Data     []byte
}

// IsEmpty returns true if the frame contains no data, false otherwise
func (frame Frame) IsEmpty() bool {
	return len(frame.Data) == 0
}

// writeFrame writes a single frame to the io.Writer
func writeFrame(output io.Writer, streamID uint16, payload []byte) error {
	if len(payload) > MaxPayloadSize {
		return ErrPayloadTooBig
	}

	var err error
	var header [4]byte
	var payloadLen = uint16(len(payload))

	// Construct the header
	binary.BigEndian.PutUint16(header[0:2], streamID)
	binary.BigEndian.PutUint16(header[2:4], payloadLen)

	// Write the header
	_, err = output.Write(header[:])
	if err != nil {
		return err
	}

	// Write the payload
	if payloadLen == 0 {
		return nil
	}
	_, err = output.Write(payload[0:payloadLen])

	return err
}

// readFrame reads a single frame from the io.Reader
func readFrame(input io.Reader) (streamID uint16, data []byte, err error) {
	var header [4]byte
	var n int

	// ReadRequest the header
	n, err = input.Read(header[:])
	if n != 4 {
		return streamID, data, ErrReadError
	}

	// Parse the header
	streamID = binary.BigEndian.Uint16(header[0:2])
	size := binary.BigEndian.Uint16(header[2:4])

	// ReadRequest the data if any
	if size > 0 {
		data = make([]byte, size)
		n, err = input.Read(data)
		if err != nil {
			return 0, nil, ErrReadError
		}
	}

	return
}
