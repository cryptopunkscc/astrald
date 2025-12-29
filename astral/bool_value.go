package astral

import (
	"encoding/binary"
	"encoding/json"
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
	err = binary.Write(w, ByteOrder, b.Bool())
	if err == nil {
		n = 1
	}
	return
}

func (b boolValue) ReadFrom(r io.Reader) (n int64, err error) {
	var v bool
	err = binary.Read(r, ByteOrder, &v)
	if err == nil {
		n = 1
		b.SetBool(v)
	}
	return
}

func (b boolValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.Bool())
}

func (b boolValue) UnmarshalJSON(bytes []byte) (err error) {
	var v bool
	err = json.Unmarshal(bytes, &v)
	b.SetBool(v)
	return
}
