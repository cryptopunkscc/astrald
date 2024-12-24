package astral

import (
	"errors"
	"io"
)

// Bytes is an unconstrained byte buffer. WriteTo will simply write the entire
// slice, ReadFrom will read until io.EOF.
type Bytes []byte

type Bytes8 []byte

type Bytes16 []byte

type Bytes32 []byte

type Bytes64 []byte

func (b Bytes) ObjectType() string {
	return "bytes"
}

func (Bytes8) ObjectType() string {
	return "bytes8"
}

func (Bytes16) ObjectType() string {
	return "bytes16"
}

func (Bytes32) ObjectType() string {
	return "bytes32"
}

func (Bytes64) ObjectType() string {
	return "bytes64"
}

func (b Bytes) WriteTo(w io.Writer) (_ int64, err error) {
	var m int
	m, err = w.Write(b)
	return int64(m), err
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

func (b *Bytes) ReadFrom(r io.Reader) (n int64, err error) {
	var buf []byte
	buf, err = io.ReadAll(r)
	n = int64(len(buf))
	if err == nil {
		*b = buf
	}
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

func init() {
	var (
		b   Bytes
		b8  Bytes8
		b16 Bytes16
		b32 Bytes32
		b64 Bytes64
	)

	DefaultBlueprints.Add(&b, &b8, &b16, &b32, &b64)
}
