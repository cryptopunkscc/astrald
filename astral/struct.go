package astral

import (
	"fmt"
	"io"
	"reflect"
)

// Struct returns an untyped pseudo-object that wraps [a pointer to] a strucutre. Its WriteTo and ReadFrom methods
// will iterate over exported fields of the wrapped value and call their respective WriteTo/ReadFrom methods.
// Returns total bytes writted/read.
func Struct(a any) Object {
	return &structWrapper{a}
}

var _ Object = &structWrapper{}

type structWrapper struct {
	s any
}

// astral:blueprint-ignore
func (s structWrapper) ObjectType() string {
	return ""
}

func (s structWrapper) WriteTo(w io.Writer) (n int64, err error) {
	val := reflect.ValueOf(s.s)

	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		err = fmt.Errorf("expected a struct, got %s", val.Kind())
		return
	}

	var m int64
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)

		// skip unexported fields
		if !field.CanInterface() {
			continue
		}

		if a, ok := field.Interface().([]Object); ok {
			m, err = WrapSlice(&a).WriteTo(w)
			n += m
			if err != nil {
				return
			}
			continue
		}

		if writerTo, ok := field.Interface().(io.WriterTo); ok {
			m, err = writerTo.WriteTo(w)
			n += m
			if err != nil {
				err = fmt.Errorf("error writing %s: %w", val.Type().Field(i).Name, err)
				return
			}
		} else {
			err = fmt.Errorf("field %s is not an io.WriterTo", val.Type().Field(i).Name)
			return
		}
	}

	return
}

func (s *structWrapper) ReadFrom(r io.Reader) (n int64, err error) {
	val := reflect.ValueOf(s.s)

	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		err = fmt.Errorf("expected a struct, got %s", val.Kind())
		return
	}

	var m int64
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)

		// skip unexported fields
		if !field.CanInterface() {
			continue
		}

		if field.Kind() == reflect.Pointer {
			field.Set(reflect.New(field.Type().Elem()))
		}

		if a, ok := field.Interface().([]Object); ok {
			m, err = WrapSlice(&a).ReadFrom(r)
			n += m
			if err != nil {
				return
			}
			field.Set(reflect.ValueOf(a))
			continue
		}

		rf, ok := field.Interface().(io.ReaderFrom)
		if !ok && field.CanAddr() {
			rf, ok = field.Addr().Interface().(io.ReaderFrom)
		}
		if !ok {
			err = fmt.Errorf("field %s is not a io.ReaderFrom", val.Type().Field(i).Name)
			return
		}

		m, err = rf.ReadFrom(r)
		n += m
		if err != nil {
			err = fmt.Errorf("error reading into field %s: %w", val.Type().Field(i).Name, err)
			return
		}
	}

	return
}
