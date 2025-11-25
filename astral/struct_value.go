package astral

import (
	"bytes"
	"io"
	"reflect"
)

type structValue struct {
	reflect.Value
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

	for i := range s.NumField() {
		var f = s.Field(i)
		if !f.CanInterface() {
			continue
		}

		if f.Kind() == reflect.Ptr {
			if f.IsNil() {
				m, err = Bytes32{}.WriteTo(w)
				n += m
				if err != nil {
					return
				}
				continue
			}

			f = f.Elem()

			if o, ok := f.Interface().(Object); ok {
				m, err = objectWrapper{o}.WriteTo(w)
				n += m
				if err != nil {
					return
				}
				continue
			}

			if f.CanAddr() {
				if o, ok := f.Addr().Interface().(Object); ok {
					m, err = objectWrapper{o}.WriteTo(w)
					n += m
					if err != nil {
						return
					}
					continue
				}
			}

		}

		if o, ok := f.Interface().(Object); ok {
			m, err = o.WriteTo(w)
			n += m
			if err != nil {
				return
			}
			continue
		}

		if f.CanAddr() {
			if o, ok := f.Addr().Interface().(Object); ok {
				m, err = o.WriteTo(w)
				n += m
				if err != nil {
					return
				}
				continue
			}
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

	for i := range s.NumField() {
		var f = s.Field(i)
		if !f.CanInterface() {
			continue
		}

		if f.Kind() == reflect.Ptr {
			var buf Bytes32
			m, err = buf.ReadFrom(r)
			n += m
			if err != nil {
				return
			}

			if len(buf) == 0 {
				f.Set(reflect.Zero(f.Type()))
				continue
			}

			f.Set(reflect.New(f.Type().Elem()))

			f = f.Elem()

			if o, ok := f.Interface().(Object); ok {
				m, err = o.ReadFrom(bytes.NewReader(buf))
				n += m
				if err != nil {
					return
				}
				continue
			}

			if f.CanAddr() {
				if o, ok := f.Addr().Interface().(Object); ok {
					m, err = o.ReadFrom(bytes.NewReader(buf))
					n += m
					if err != nil {
						return
					}
					continue
				}
			}

		}

		if o, ok := f.Interface().(Object); ok {
			m, err = o.ReadFrom(r)
			n += m
			continue
		}

		if f.CanAddr() {
			if o, ok := f.Addr().Interface().(Object); ok {
				m, err = o.ReadFrom(r)
				n += m
				continue
			}
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
