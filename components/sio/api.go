package sio

import "io"

type ReadWriteCloser interface {
	io.Closer
	ReadWrite
}

type ReadWrite interface {
	Reader
	Writer
}

type ReadCloser interface {
	io.ReadCloser
	Reader
}

type Reader interface {
	io.Reader
	Deserializer
}

type WriteCloser interface {
	io.Closer
	Writer
}

type Writer interface {
	io.Writer
	Serializer
}

type Deserializer interface {
	ReadByte() (byte, error)
	ReadUint8() (uint8, error)
	ReadUint16() (uint16, error)
	ReadUint32() (uint32, error)
	ReadUint64() (uint64, error)
	ReadWithSize8() (buff []byte, err error)
	ReadWithSize16() (buff []byte, err error)
	ReadWithSize32() (buff []byte, err error)
	ReadN(n int) ([]byte, error)
	ReadStringWithSize8() (string, error)
	ReadStringWithSize16() (string, error)
	ReadStringWithSize32() (string, error)
	ReadString(n int) (string, error)
}

type Serializer interface {
	WriteByte(b byte) (err error)
	WriteUInt8(i uint8) error
	WriteUInt16(i uint16) (int, error)
	WriteUInt32(i uint32) (int, error)
	WriteUInt64(i uint64) (int, error)
	WriteString(s string) (int, error)
	WriteWithSize8(b []byte) (l int, err error)
	WriteWithSize16(b []byte) (l int, err error)
	WriteWithSize32(b []byte) (l int, err error)
	WriteStringWithSize8(s string) (int, error)
	WriteStringWithSize16(s string) (int, error)
	WriteStringWithSize32(s string) (int, error)
}
