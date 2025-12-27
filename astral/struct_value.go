package astral

import (
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strings"
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

	// use struct's own WriteTo only if it's not a root element to avoid infinite loops
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
	// if the struct is non-root try to cast it to a json.Marshaler
	if !s.root {
		o, ok := s.Interface().(json.Marshaler)
		if ok {
			return o.MarshalJSON()
		}
		if s.CanAddr() {
			o, ok = s.Addr().Interface().(json.Marshaler)
			if ok {
				return o.MarshalJSON()
			}
		}
	}

	var v = map[string]json.RawMessage{}

	for i := range s.NumField() {
		f := s.Field(i)

		// skip unexported fields
		if !f.CanInterface() {
			continue
		}

		fobject, err := objectify(f)
		if err != nil {
			return nil, err
		}

		fname := s.Type().Field(i).Name

		v[fname], err = fobject.MarshalJSON()
		if err != nil {
			return nil, err
		}
	}

	return json.Marshal(v)
}

func (s structValue) UnmarshalJSON(data []byte) error {
	// if the struct is non-root try to cast it to a json.Unmarshaler
	if !s.root {
		if o, ok := s.Interface().(json.Unmarshaler); ok {
			return o.UnmarshalJSON(data)
		}
		if s.CanAddr() {
			o, ok := s.Addr().Interface().(json.Unmarshaler)
			if ok {
				return o.UnmarshalJSON(data)
			}
		}
	}

	var fields map[string]json.RawMessage

	err := json.Unmarshal(data, &fields)
	if err != nil {
		return err
	}

	// convert all keys to lowercase
	for k, v := range fields {
		l := strings.ToLower(k)
		if k == l {
			continue
		}
		if _, dup := fields[l]; dup {
			return errors.New("object has duplicate fields due to case insensitivity")
		}
		fields[l] = v
		delete(fields, k)
	}

	for i := range s.NumField() {
		f := s.Field(i)

		// skip unexported fields
		if !f.CanInterface() {
			continue
		}

		fname := strings.ToLower(s.Type().Field(i).Name)

		jdata, ok := fields[fname]
		if !ok {
			continue
		}

		fobject, err := objectify(f)
		if err != nil {
			return err
		}

		err = fobject.UnmarshalJSON(jdata)
		if err != nil {
			return err
		}

		delete(fields, fname)
	}

	if len(fields) > 0 {
		return errors.New("excess fields in json object")
	}

	return nil
}
