package sio

import (
	"bytes"
	"encoding/binary"
	"io"
)

type writer struct {
	io.Writer
	io.Closer
}

func (f *writer) WriteByte(b byte) (err error) {
	_, err = f.Write([]byte{b})
	return
}

func (f *writer) WriteUInt8(i uint8) error {
	return f.WriteByte(i)
}

func (f *writer) WriteUInt16(i uint16) (int, error) {
	var buff [2]byte
	binary.BigEndian.PutUint16(buff[:], i)
	return f.Write(buff[:])
}

func (f *writer) WriteUInt32(i uint32) (int, error) {
	var buff [4]byte
	binary.BigEndian.PutUint32(buff[:], i)
	return f.Write(buff[:])
}

func (f *writer) WriteUInt64(i uint64) (int, error) {
	var buff [8]byte
	binary.BigEndian.PutUint64(buff[:], i)
	return f.Write(buff[:])
}

func (f *writer) WriteString(s string) (int, error) {
	return f.Write(bytes.NewBufferString(s).Bytes())
}

func (f *writer) WriteWithSize8(b []byte) (l int, err error) {
	err = f.WriteUInt8(uint8(len(b)))
	if err != nil {
		return
	}
	l2, err := f.Write(b)
	l = l + l2
	return
}

func (f *writer) WriteWithSize16(b []byte) (l int, err error) {
	_, err = f.WriteUInt16(uint16(len(b)))
	if err != nil {
		return
	}
	l2, err := f.Write(b)
	l = l + l2
	return
}

func (f *writer) WriteWithSize32(b []byte) (l int, err error) {
	_, err = f.WriteUInt32(uint32(len(b)))
	if err != nil {
		return
	}
	l2, err := f.Write(b)
	l = l + l2
	return
}

func (f *writer) WriteStringWithSize8(s string) (int, error) {
	return f.WriteWithSize8(bytes.NewBufferString(s).Bytes())
}

func (f *writer) WriteStringWithSize16(s string) (int, error) {
	return f.WriteWithSize16(bytes.NewBufferString(s).Bytes())
}

func (f *writer) WriteStringWithSize32(s string) (int, error) {
	return f.WriteWithSize32(bytes.NewBufferString(s).Bytes())
}
