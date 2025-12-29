package astral

import (
	"bytes"
	"encoding/binary"
	"io"
)

// objectWrapper is an object that encodes the payload of the wrapped object using a 32-bit length prefix.
type objectWrapper struct {
	o Object
}

var _ Object = &objectWrapper{}

// ObjectType is always implicit
func (objectWrapper) ObjectType() string { return "" }

func (o objectWrapper) WriteTo(w io.Writer) (n int64, err error) {
	if o.o == nil {
		err = binary.Write(w, ByteOrder, uint32(0))
		if err == nil {
			n += 4
		}
		return
	}

	var buf = &bytes.Buffer{}
	var m int64

	_, err = o.o.WriteTo(buf)
	if err != nil {
		return
	}

	err = binary.Write(w, ByteOrder, uint32(buf.Len()))
	if err != nil {
		return
	}
	n = 4

	m, err = buf.WriteTo(w)
	n += m

	return
}

func (o objectWrapper) ReadFrom(r io.Reader) (n int64, err error) {
	var l uint32
	var m int
	var k int64
	err = binary.Read(r, ByteOrder, &l)
	if err != nil {
		return
	}
	n += 4

	var buf = make([]byte, l)
	m, err = io.ReadFull(r, buf)
	n += int64(m)
	if err != nil {
		return
	}

	k, err = o.o.ReadFrom(bytes.NewReader(buf))
	n += k

	return
}
