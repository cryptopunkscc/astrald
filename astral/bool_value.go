package astral

import (
	"encoding/binary"
	"io"
	"reflect"
)

type boolValue struct {
	reflect.Value
}

var _ Object = &boolValue{}

func (b boolValue) ObjectType() string {
	return "bool"
}

func (b boolValue) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, encoding, b.Bool())
	if err == nil {
		n = 1
	}
	return
}

func (b boolValue) ReadFrom(r io.Reader) (n int64, err error) {
	var v bool
	err = binary.Read(r, encoding, &v)
	if err == nil {
		n = 1
	}
	b.Set(reflect.ValueOf(v))
	return
}
