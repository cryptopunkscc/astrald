package serializer

import (
	"bytes"
	"encoding/binary"
	"io"
)

type Writer struct {
	io.Writer
}

func (f Writer) WriteByte(b byte) (err error) {
	_, err = f.Write([]byte{b})
	return
}

func (f Writer) WriteWithSize(b []byte) (l int, err error) {
	l, err = f.WriteSize(len(b))
	if err != nil {
		return
	}
	l2, err := f.Write(b)
	l = l + l2
	return
}

func (f Writer) WriteString(s string) (int, error) {
	return f.Write(bytes.NewBufferString(s).Bytes())
}

func (f Writer) WriteStringWithSize(s string) (int, error) {
	buff := bytes.NewBufferString(s).Bytes()
	return f.WriteWithSize(buff)
}

func (f Writer) WriteSize(i int) (int, error) {
	return f.WriteInt(i)
}

func (f Writer) WriteInt(i int) (int, error) {
	var buff [4]byte
	binary.BigEndian.PutUint32(buff[:], uint32(i))
	return f.Write(buff[:])
}

func (f Writer) WriteUInt16(i uint16) (int, error) {
	var buff [2]byte
	binary.BigEndian.PutUint16(buff[:], i)
	return f.Write(buff[:])
}

func (f Writer) WriteUInt32(i uint32) (int, error) {
	var buff [4]byte
	binary.BigEndian.PutUint32(buff[:], i)
	return f.Write(buff[:])
}

func (f Writer) WriteUInt64(i uint64) (int, error) {
	var buff [8]byte
	binary.BigEndian.PutUint64(buff[:], i)
	return f.Write(buff[:])
}
