package astral

import (
	"encoding/binary"
	"io"
	"reflect"
)

type mapValue struct {
	reflect.Value
}

var _ Object = &mapValue{}

func (m mapValue) ObjectType() string {
	return ""
}

func (m mapValue) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, encoding, uint32(m.Len()))
	if err != nil {
		return
	}
	n += 4

	var o Object
	var i int64

	for _, k := range m.MapKeys() {
		o, err = objectify(k)
		if err != nil {
			return
		}

		i, err = objectWrapper{o}.WriteTo(w)
		n += i
		if err != nil {
			return
		}

		o, err = objectify(m.MapIndex(k))
		if err != nil {
			return
		}

		i, err = objectWrapper{o}.WriteTo(w)
		n += i
		if err != nil {
			return
		}
	}

	return
}

func (m mapValue) ReadFrom(r io.Reader) (n int64, err error) {
	var l uint32
	err = binary.Read(r, encoding, &l)
	if err != nil {
		return
	}
	n += 4

	if l == 0 {
		reflect.Zero(m.Type())
		return
	}

	m.Set(reflect.MakeMap(m.Type()))

	var o Object
	var k int64

	for range l {
		var key = reflect.New(m.Type().Key()).Elem()

		o, err = objectify(key)
		k, err = objectWrapper{o}.ReadFrom(r)
		n += k
		if err != nil {
			return
		}

		var value = reflect.New(m.Type().Elem()).Elem()
		o, err = objectify(value)
		k, err = objectWrapper{o}.ReadFrom(r)
		n += k
		if err != nil {
			return
		}

		m.SetMapIndex(key, value)
	}

	return
}

func (m mapValue) MarshalJSON() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (m mapValue) UnmarshalJSON(bytes []byte) error {
	//TODO implement me
	panic("implement me")
}
