package astral

import (
	"bytes"
	"encoding/binary"
	"io"
	"reflect"
)

type ptrValue struct {
	reflect.Value
}

var _ Object = &ptrValue{}

func (p ptrValue) ObjectType() string {
	return ""
}

func (p ptrValue) WriteTo(w io.Writer) (n int64, err error) {
	var buf = &bytes.Buffer{}
	var o Object

	if p.IsNil() {
		err = binary.Write(w, encoding, uint32(0))
		if err == nil {
			n += 4
		}
		return
	}

	o, err = objectify(p.Elem())
	if err != nil {
		return 0, err
	}

	n, err = o.WriteTo(buf)
	if err != nil {
		return 0, err
	}

	err = binary.Write(w, encoding, uint32(buf.Len()))
	if err != nil {
		return 0, err
	}
	n = 4

	var m int
	m, err = w.Write(buf.Bytes())

	n += int64(m)
	return
}

func (p ptrValue) ReadFrom(r io.Reader) (n int64, err error) {
	// read the length
	var l uint32
	err = binary.Read(r, encoding, &l)
	if err != nil {
		return
	}
	n += 4

	// zero length means nil
	if l == 0 {
		p.Set(reflect.Zero(p.Type()))
		return
	}

	// read the data
	var buf = make([]byte, l)
	var m int
	m, err = io.ReadFull(r, buf)
	n += int64(m)
	if err != nil {
		return
	}

	// initialize the element
	p.Set(reflect.New(p.Type().Elem()))

	var o Object
	o, err = objectify(p.Elem())
	if err != nil {
		return 0, err
	}

	// read the data
	var k int64
	k, err = o.ReadFrom(bytes.NewReader(buf))
	n += k

	return
}
