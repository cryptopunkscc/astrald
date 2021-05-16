package proto

import (
	"encoding/binary"
	"errors"
)

// Request represents a request to open a port
type Request struct {
	StreamID uint16 // Sender's local stream ID for receiving the response
	Port     string // Receiver's port that the sender wants to open
}

var ErrParseError = errors.New("parse error")

// Bytes returns a byte representation of a Request
func (msg Request) Bytes() []byte {
	msgLen := len(msg.Port) + 2
	bytes := make([]byte, msgLen)

	binary.BigEndian.PutUint16(bytes[0:2], msg.StreamID)
	copy(bytes[2:], msg.Port)

	return bytes
}

// ParseRequest parses the byte buffer for a Request message
func ParseRequest(data []byte) (Request, error) {
	// Basic validity check
	if len(data) < 2 {
		return Request{}, ErrParseError
	}

	// Parse data
	remoteStreamID := binary.BigEndian.Uint16(data[0:2])
	port := string(data[2:])

	return Request{
		StreamID: remoteStreamID,
		Port:     port,
	}, nil
}
