package astral

import (
	"encoding/json"
	"errors"
	"io"
	"reflect"
)

type Objectified struct {
	reflect.Value
}

// Objectify converts any value into an Object. If the given value is a struct with an ObjectType() method,
// that method will be used as the returned object's type; otherwise the type is empty.
// The argument must be a pointer to a variable for both reading and writing.
func Objectify(a any) Objectified {
	var v = reflect.ValueOf(a)

	switch {
	case v.Kind() != reflect.Ptr:
		panic("argument must be a pointer")
	case v.IsNil():
		panic("cannot objectify nil pointer")
	}

	return Objectified{Value: v}
}

// astral:blueprint-ignore
func (o Objectified) ObjectType() string {
	elem, err := objectify(o.Elem())
	if err != nil {
		return ""
	}

	return elem.ObjectType()
}

func (o Objectified) WriteTo(w io.Writer) (n int64, err error) {
	if o.Elem().Kind() == reflect.Struct {
		return structValue{
			Value: o.Elem(),
			root:  true,
		}.WriteTo(w)
	}

	e, err := objectify(o.Elem())
	if err != nil {
		return 0, err
	}

	return e.WriteTo(w)
}

func (o Objectified) ReadFrom(r io.Reader) (n int64, err error) {
	if o.Elem().Kind() == reflect.Struct {
		return structValue{
			Value: o.Elem(),
			root:  true,
		}.ReadFrom(r)
	}

	e, err := objectify(o.Elem())
	if err != nil {
		return 0, err
	}
	return e.ReadFrom(r)
}

func (o Objectified) MarshalJSON() ([]byte, error) {
	if o.Elem().Kind() == reflect.Struct {
		return structValue{
			Value: o.Elem(),
			root:  true,
		}.MarshalJSON()
	}

	e, err := objectify(o.Elem())
	if err != nil {
		return nil, err
	}

	return e.MarshalJSON()
}

func (o Objectified) UnmarshalJSON(bytes []byte) error {
	if o.Elem().Kind() == reflect.Struct {
		return structValue{
			Value: o.Elem(),
			root:  true,
		}.UnmarshalJSON(bytes)
	}

	e, err := objectify(o.Elem())
	if err != nil {
		return err
	}
	return e.UnmarshalJSON(bytes)
}

// value is an interface for objects that support both binary and JSON encoding.
type value interface {
	Object
	json.Marshaler
	json.Unmarshaler
}

func objectify(v reflect.Value) (value, error) {
	switch v.Kind() {
	case reflect.Interface, reflect.Ptr:
	default:
		if v.CanInterface() {
			if v, ok := v.Interface().(value); ok {
				return v, nil
			}
		}
		if v.CanAddr() {
			if v, ok := v.Addr().Interface().(value); ok {
				return v, nil
			}
		}
	}

	switch v.Kind() {
	case reflect.Ptr:
		return ptrValue{Value: v}, nil

	case reflect.Uint8:
		return uint8Value{v}, nil

	case reflect.Uint16:
		return uint16Value{v}, nil

	case reflect.Uint32:
		return uint32Value{v}, nil

	case reflect.Uint64, reflect.Uint:
		return uint64Value{v}, nil

	case reflect.Int8:
		return int8Value{v}, nil

	case reflect.Int16:
		return int16Value{v}, nil

	case reflect.Int32:
		return int32Value{v}, nil

	case reflect.Int64, reflect.Int:
		return int64Value{v}, nil

	case reflect.Array:
		return arrayValue{v}, nil

	case reflect.Slice:
		return sliceValue{v}, nil

	case reflect.String:
		return stringValue{v}, nil

	case reflect.Map:
		return mapValue{v}, nil

	case reflect.Struct:
		return structValue{v, false}, nil

	case reflect.Interface:
		return interfaceValue{v}, nil

	case reflect.Bool:
		return boolValue{v}, nil

	default:
		return nil, errors.New("unsupported type " + v.Kind().String() + " " + v.Type().String())
	}
}
