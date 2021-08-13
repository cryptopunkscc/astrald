package proto

import (
	"encoding/binary"
	"errors"
)

var ErrParseError = errors.New("parse error")

func MakeQuery(stream int, port string) []byte {
	msgLen := len(port) + 2
	bytes := make([]byte, msgLen)

	binary.BigEndian.PutUint16(bytes[0:2], uint16(stream))
	copy(bytes[2:], port)

	return bytes
}

func ParseQuery(data []byte) (stream int, port string, err error) {
	if len(data) < 2 {
		return 0, "", ErrParseError
	}

	stream = int(binary.BigEndian.Uint16(data[0:2]))
	port = string(data[2:])

	return
}
