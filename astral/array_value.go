package astral

import (
	"io"
	"reflect"
)

type arrayValue struct {
	reflect.Value
}

var _ Object = &arrayValue{}

func (a arrayValue) ObjectType() string {
	return ""
}

func (a arrayValue) WriteTo(w io.Writer) (n int64, err error) {
	var o Object
	var m int64

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

func (a arrayValue) ReadFrom(r io.Reader) (n int64, err error) {
	var o Object
	var m int64

	for i := range a.Len() {
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

func (a arrayValue) MarshalJSON() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (a arrayValue) UnmarshalJSON(bytes []byte) error {
	//TODO implement me
	panic("implement me")
}
