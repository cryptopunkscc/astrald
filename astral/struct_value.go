package astral

import (
	"encoding/json"
	"io"
	"reflect"
)

type structValue struct {
	reflect.Value
	root bool
}

var _ Object = &structValue{}

func (s structValue) ObjectType() string {
	if t, ok := s.Interface().(ObjectTyper); ok {
		return t.ObjectType()
	}

	return ""
}

func (s structValue) WriteTo(w io.Writer) (n int64, err error) {
	var m int64
	var o Object

	// if the struct is non-root and an object, use its own WriteTo
	if !s.root {
		o, ok := s.Interface().(Object)
		if ok {
			return o.WriteTo(w)
		}
		if s.CanAddr() {
			o, ok = s.Addr().Interface().(Object)
			if ok {
				return o.WriteTo(w)
			}
		}
	}

	for i := range s.NumField() {
		var f = s.Field(i)
		if !f.CanInterface() {
			continue
		}

		o, err = objectify(f)
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

func (s structValue) ReadFrom(r io.Reader) (n int64, err error) {
	var m int64
	var o Object

	// if the struct is non-root and an object, use its own ReadFrom
	if !s.root {
		if o, ok := s.Interface().(Object); ok {
			return o.ReadFrom(r)
		}
		if s.CanAddr() {
			o, ok := s.Addr().Interface().(Object)
			if ok {
				return o.ReadFrom(r)
			}
		}
	}

	for i := range s.NumField() {
		var f = s.Field(i)
		if !f.CanInterface() {
			continue
		}

		o, err = objectify(f)
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

func (s structValue) MarshalJSON() ([]byte, error) {
	var v = map[string]any{}

	for i := range s.NumField() {
		f := s.Field(i)

		// skip unexported fields
		if !f.CanInterface() {
			continue
		}

		v[s.Type().Field(i).Name] = s.Field(i).Interface()
	}

	return json.Marshal(v)
}

func (s structValue) UnmarshalJSON(i []byte) error {
	//TODO implement me
	panic("implement me")
}
