package serialize

import (
	"bytes"
	"encoding/binary"
	"io"
)

type Formatter struct {
	io.Writer
}

func NewFormatter(writer io.Writer) *Formatter {
	return &Formatter{Writer: writer}
}

func (f Formatter) WriteByte(b byte) (err error) {
	_, err = f.Write([]byte{b})
	return
}

func (f Formatter) WriteWithSize(b []byte) (l int, err error) {
	l, err = f.WriteSize(len(b))
	if err != nil {
		return
	}
	l2, err := f.Write(b)
	l = l + l2
	return
}

func (f Formatter) WriteString(s string) (int, error) {
	return f.Write(bytes.NewBufferString(s).Bytes())
}

func (f Formatter) WriteStringWithSize(s string) (int, error) {
	buff := bytes.NewBufferString(s).Bytes()
	return f.WriteWithSize(buff)
}

func (f Formatter) WriteSize(i int) (int, error) {
	return f.WriteInt(i)
}

func (f Formatter) WriteInt(i int) (int, error) {
	var buff [4]byte
	binary.BigEndian.PutUint32(buff[:], uint32(i))
	return f.Write(buff[:])
}

func (f Formatter) WriteUInt16(i uint16) (int, error) {
	var buff [2]byte
	binary.BigEndian.PutUint16(buff[:], i)
	return f.Write(buff[:])
}

func (f Formatter) WriteUInt32(i uint32) (int, error) {
	var buff [4]byte
	binary.BigEndian.PutUint32(buff[:], i)
	return f.Write(buff[:])
}

func (f Formatter) WriteUInt64(i uint64) (int, error) {
	var buff [8]byte
	binary.BigEndian.PutUint64(buff[:], i)
	return f.Write(buff[:])
}
