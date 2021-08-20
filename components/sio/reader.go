package sio

import (
	"encoding/binary"
	"io"
)

type reader struct {
	io.Reader
	io.Closer
}

func (r *reader) ReadByte() (byte, error) {
	buff, err := r.ReadN(1)
	if err != nil {
		return 0, err
	}
	return buff[0], nil
}

func (r *reader) ReadUint8() (uint8, error) {
	b, err := r.ReadByte()
	if err != nil {
		return 0, err
	}
	return b, nil
}

func (r *reader) ReadUint16() (uint16, error) {
	buff, err := r.ReadN(2)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(buff[:]), nil
}

func (r *reader) ReadUint32() (uint32, error) {
	buff, err := r.ReadN(4)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(buff[:]), nil
}

func (r *reader) ReadUint64() (uint64, error) {
	buff, err := r.ReadN(8)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(buff[:]), nil
}

func (r *reader) ReadN(n int) ([]byte, error) {
	buff := make([]byte, n)
	read, err := r.Read(buff)
	if err != nil {
		return nil, err
	}
	return buff[:read], nil
}

func (r *reader) ReadWithSize8() (buff []byte, err error) {
	size, err := r.ReadUint8()
	if err != nil {
		return
	}
	buff, err = r.ReadN(int(size))
	if err != nil {
		return
	}
	return
}

func (r *reader) ReadWithSize16() (buff []byte, err error) {
	size, err := r.ReadUint16()
	if err != nil {
		return
	}
	buff, err = r.ReadN(int(size))
	if err != nil {
		return
	}
	return
}

func (r *reader) ReadWithSize32() (buff []byte, err error) {
	size, err := r.ReadUint32()
	if err != nil {
		return
	}
	buff, err = r.ReadN(int(size))
	if err != nil {
		return
	}
	return
}

func (r *reader) ReadString(n int) (string, error) {
	buff, err := r.ReadN(n)
	if err != nil {
		return "", err
	}
	return string(buff[:]), nil
}

func (r *reader) ReadStringWithSize8() (string, error) {
	buff, err := r.ReadWithSize8()
	if err != nil {
		return "", err
	}
	return string(buff), err
}

func (r *reader) ReadStringWithSize16() (string, error) {
	buff, err := r.ReadWithSize16()
	if err != nil {
		return "", err
	}
	return string(buff), err
}

func (r *reader) ReadStringWithSize32() (string, error) {
	buff, err := r.ReadWithSize32()
	if err != nil {
		return "", err
	}
	return string(buff), err
}
