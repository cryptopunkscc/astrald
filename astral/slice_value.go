package astral

import (
	"encoding/binary"
	"io"
	"reflect"
)

type sliceValue struct {
	reflect.Value
}

var _ Object = &sliceValue{}

func (a sliceValue) ObjectType() string {
	return ""
}

func (a sliceValue) WriteTo(w io.Writer) (n int64, err error) {
	var o Object
	var m int64

	err = binary.Write(w, encoding, uint32(a.Len()))
	if err != nil {
		return
	}
	n += 4

	for i := range a.Len() {
		o, err = objectify(a.Index(i))
		if err != nil {
			return
		}
		m, err = o.WriteTo(w)
		n += m
		if err != nil {
			return
		}
	}
	return
}

func (a sliceValue) ReadFrom(r io.Reader) (n int64, err error) {
	var o Object
	var m int64
	var l uint32

	err = binary.Read(r, encoding, &l)
	if err != nil {
		return
	}

	a.Set(reflect.MakeSlice(a.Type(), int(l), int(l)))

	for i := range int(l) {
		o, err = objectify(a.Index(i))
		if err != nil {
			return
		}
		m, err = o.ReadFrom(r)
		n += m
		if err != nil {
			return
		}
	}
	return
}
