package astral

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"reflect"
)

type sliceValue struct {
	reflect.Value
}

var _ Object = &sliceValue{}

// astral:blueprint-ignores
func (a sliceValue) ObjectType() string {
	return ""
}

func (a sliceValue) WriteTo(w io.Writer) (n int64, err error) {
	var o Object
	var m int64

	err = binary.Write(w, ByteOrder, uint32(a.Len()))
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

	err = binary.Read(r, ByteOrder, &l)
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

func (a sliceValue) MarshalJSON() (_ []byte, err error) {
	var arr []json.RawMessage
	var o Object
	var raw []byte

	for i := range a.Len() {
		o, err = objectify(a.Index(i))
		if err != nil {
			return
		}

		m, ok := o.(json.Marshaler)
		if !ok {
			return nil, errors.New("object does not implement json encoding")
		}

		raw, err = m.MarshalJSON()
		if err != nil {
			return
		}

		arr = append(arr, raw)
	}

	return json.Marshal(arr)
}

func (a sliceValue) UnmarshalJSON(bytes []byte) error {
	var arr []json.RawMessage

	err := json.Unmarshal(bytes, &arr)
	if err != nil {
		return err
	}

	a.Set(reflect.MakeSlice(a.Type(), len(arr), len(arr)))

	for i, raw := range arr {
		o, err := objectify(a.Index(i))
		if err != nil {
			return err
		}

		m, ok := o.(json.Unmarshaler)
		if !ok {
			return errors.New("object does not implement json encoding")
		}

		err = m.UnmarshalJSON(raw)
		if err != nil {
			return err
		}
	}

	return nil
}
