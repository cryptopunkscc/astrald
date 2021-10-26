package proto

import (
	"encoding/binary"
	"errors"
)

var ErrParseError = errors.New("parse error")

func MakeQuery(stream int, query string) []byte {
	msgLen := len(query) + 2
	bytes := make([]byte, msgLen)

	binary.BigEndian.PutUint16(bytes[0:2], uint16(stream))
	copy(bytes[2:], query)

	return bytes
}

func ParseQuery(data []byte) (stream int, query string, err error) {
	if len(data) < 2 {
		return 0, "", ErrParseError
	}

	stream = int(binary.BigEndian.Uint16(data[0:2]))
	query = string(data[2:])

	return
}
