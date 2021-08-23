package binary

import (
	"encoding/binary"
)

func Int8UBytes(i int8) []byte {
	return []byte{uint8(i)}
}

func Int16UBytes(i int16) []byte {
	var bytes [2]byte
	binary.BigEndian.PutUint16(bytes[:], uint16(i))
	return bytes[:]
}

func Int32UBytes(i int32) []byte {
	var bytes [4]byte
	binary.BigEndian.PutUint32(bytes[:], uint32(i))
	return bytes[:]
}

func Int64UBytes(i int64) []byte {
	var bytes [8]byte
	binary.BigEndian.PutUint64(bytes[:], uint64(i))
	return bytes[:]
}

func UBytesInt8(bytes []byte) (i int8) {
	return int8(bytes[0])
}

func UBytesInt16(bytes []byte) (i int16) {
	return int16(binary.BigEndian.Uint16(bytes))
}

func UBytesInt32(bytes []byte) (i int32) {
	return int32(binary.BigEndian.Uint32(bytes))
}

func UBytesInt64(bytes []byte) (i int64) {
	return int64(binary.BigEndian.Uint64(bytes))
}