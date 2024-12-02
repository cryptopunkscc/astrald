package astral

import (
	"encoding/binary"
	"errors"
	"io"
)

var encoding = binary.BigEndian

type Bytes64 []byte

type Bytes32 []byte

type Bytes16 []byte

type Bytes8 []byte

type String64 string

type String32 string

type String16 string

type String8 string

type Uint64 uint64

type Uint32 uint32

type Uint16 uint16

type Uint8 uint8

type Int64 int64

type Int32 int32

type Int16 int16

type Int8 int8

func (Bytes64) ObjectType() string {
	return "bytes64"
}

func (Bytes32) ObjectType() string {
	return "bytes32"
}

func (Bytes16) ObjectType() string {
	return "bytes16"
}

func (Bytes8) ObjectType() string {
	return "bytes8"
}

func (String64) ObjectType() string {
	return "string64"
}

func (String32) ObjectType() string {
	return "string32"
}

func (String16) ObjectType() string {
	return "string16"
}

func (String8) ObjectType() string {
	return "string8"
}

func (Uint64) ObjectType() string {
	return "uint64"
}

func (Uint32) ObjectType() string {
	return "uint32"
}

func (Uint16) ObjectType() string {
	return "uint16"
}

func (Uint8) ObjectType() string {
	return "uint8"
}

func (Int64) ObjectType() string {
	return "int64"
}

func (Int32) ObjectType() string {
	return "int32"
}

func (Int16) ObjectType() string {
	return "int16"
}

func (Int8) ObjectType() string {
	return "int8"
}

func (b Bytes64) WriteTo(w io.Writer) (n int64, err error) {
	var l = Uint64(len(b))

	n, err = l.WriteTo(w)
	if err != nil {
		return
	}

	m, err := w.Write(b)
	n += int64(m)

	return
}

func (b Bytes32) WriteTo(w io.Writer) (n int64, err error) {
	var l = Uint32(len(b))
	if l > (1<<32)-1 {
		return 0, errors.New("data too large")
	}

	n, err = l.WriteTo(w)
	if err != nil {
		return
	}

	m, err := w.Write(b)
	n += int64(m)

	return
}

func (b Bytes16) WriteTo(w io.Writer) (n int64, err error) {
	var l = Uint16(len(b))
	if l > (1<<16)-1 {
		return 0, errors.New("data too large")
	}

	n, err = l.WriteTo(w)
	if err != nil {
		return
	}

	m, err := w.Write(b)
	n += int64(m)

	return
}

func (b Bytes8) WriteTo(w io.Writer) (n int64, err error) {
	var l = Uint8(len(b))
	if l > (1<<8)-1 {
		return 0, errors.New("data too large")
	}

	n, err = l.WriteTo(w)
	if err != nil {
		return
	}

	m, err := w.Write(b)
	n += int64(m)

	return
}

func (s String64) WriteTo(w io.Writer) (n int64, err error) {
	var l = Uint64(len(s))

	n, err = l.WriteTo(w)
	if err != nil {
		return
	}

	m, err := w.Write([]byte(s))
	n += int64(m)

	return
}

func (s String32) WriteTo(w io.Writer) (n int64, err error) {
	var l = Uint32(len(s))
	if l > (1<<32)-1 {
		return 0, errors.New("data too large")
	}

	n, err = l.WriteTo(w)
	if err != nil {
		return
	}

	m, err := w.Write([]byte(s))
	n += int64(m)

	return
}

func (s String16) WriteTo(w io.Writer) (n int64, err error) {
	var l = Uint16(len(s))
	if l > (1<<16)-1 {
		return 0, errors.New("data too large")
	}

	n, err = l.WriteTo(w)
	if err != nil {
		return
	}

	m, err := w.Write([]byte(s))
	n += int64(m)

	return
}

func (s String8) WriteTo(w io.Writer) (n int64, err error) {
	var l = Uint8(len(s))
	if l > (1<<8)-1 {
		return 0, errors.New("data too large")
	}

	n, err = l.WriteTo(w)
	if err != nil {
		return
	}

	m, err := w.Write([]byte(s))
	n += int64(m)

	return
}

func (u Uint64) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, encoding, u)
	if err == nil {
		n = 8
	}
	return
}

func (u Uint32) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, encoding, u)
	if err == nil {
		n = 4
	}
	return
}

func (u Uint16) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, encoding, u)
	if err == nil {
		n = 2
	}
	return
}

func (u Uint8) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, encoding, u)
	if err == nil {
		n = 1
	}
	return
}

func (i Int64) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, encoding, i)
	if err == nil {
		n = 8
	}
	return
}

func (i Int32) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, encoding, i)
	if err == nil {
		n = 4
	}
	return
}

func (i Int16) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, encoding, i)
	if err == nil {
		n = 2
	}
	return
}

func (i Int8) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, encoding, i)
	if err == nil {
		n = 1
	}
	return
}

func (b *Bytes64) ReadFrom(r io.Reader) (n int64, err error) {
	var l Uint64
	n, err = l.ReadFrom(r)
	if err != nil {
		return
	}

	var buf = make([]byte, l)
	m, err := io.ReadFull(r, buf)
	n += int64(m)

	*b = Bytes64(buf[:m])

	return
}

func (b *Bytes32) ReadFrom(r io.Reader) (n int64, err error) {
	var l Uint32
	n, err = l.ReadFrom(r)
	if err != nil {
		return
	}

	var buf = make([]byte, l)
	m, err := io.ReadFull(r, buf)
	n += int64(m)

	*b = Bytes32(buf[:m])

	return
}

func (b *Bytes16) ReadFrom(r io.Reader) (n int64, err error) {
	var l Uint16
	n, err = l.ReadFrom(r)
	if err != nil {
		return
	}

	var buf = make([]byte, l)
	m, err := io.ReadFull(r, buf)
	n += int64(m)

	*b = Bytes16(buf[:m])

	return
}

func (b *Bytes8) ReadFrom(r io.Reader) (n int64, err error) {
	var l Uint8
	n, err = l.ReadFrom(r)
	if err != nil {
		return
	}

	var buf = make([]byte, l)
	m, err := io.ReadFull(r, buf)
	n += int64(m)

	*b = Bytes8(buf[:m])

	return
}

func (s *String64) ReadFrom(r io.Reader) (n int64, err error) {
	var l Uint64
	n, err = l.ReadFrom(r)
	if err != nil {
		return
	}

	var buf = make([]byte, l)
	m, err := io.ReadFull(r, buf)
	n += int64(m)

	*s = String64(buf[:m])

	return
}

func (s *String32) ReadFrom(r io.Reader) (n int64, err error) {
	var l Uint32
	n, err = l.ReadFrom(r)
	if err != nil {
		return
	}

	var buf = make([]byte, l)
	m, err := io.ReadFull(r, buf)
	n += int64(m)

	*s = String32(buf[:m])

	return
}

func (s *String16) ReadFrom(r io.Reader) (n int64, err error) {
	var l Uint16
	n, err = l.ReadFrom(r)
	if err != nil {
		return
	}

	var buf = make([]byte, l)
	m, err := io.ReadFull(r, buf)
	n += int64(m)

	*s = String16(buf[:m])

	return
}

func (s *String8) ReadFrom(r io.Reader) (n int64, err error) {
	var l Uint8
	n, err = l.ReadFrom(r)
	if err != nil {
		return
	}

	var buf = make([]byte, l)
	m, err := io.ReadFull(r, buf)
	n += int64(m)

	*s = String8(buf[:m])

	return
}

func (u *Uint64) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, encoding, u)
	if err == nil {
		n = 8
	}
	return
}

func (u *Uint32) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, encoding, u)
	if err == nil {
		n = 4
	}
	return
}

func (u *Uint16) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, encoding, u)
	if err == nil {
		n = 2
	}
	return
}

func (u *Uint8) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, encoding, u)
	if err == nil {
		n = 1
	}
	return
}

func (i *Int64) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, encoding, i)
	if err == nil {
		n = 8
	}
	return
}

func (i *Int32) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, encoding, i)
	if err == nil {
		n = 4
	}
	return
}

func (i *Int16) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, encoding, i)
	if err == nil {
		n = 2
	}
	return
}

func (i *Int8) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, encoding, i)
	if err == nil {
		n = 1
	}
	return
}
