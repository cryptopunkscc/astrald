package astral

import (
	"encoding/json"
	"errors"
	"io"
	"reflect"
)

type Objectified struct {
	value
	err error
}

// Objectify converts any value into an Object. If the given value is a struct with an ObjectType() method,
// that method will be used as the returned object's type; otherwise the type is empty.
// The argument must be a pointer.
func Objectify(a any) Objectified {
	var v = reflect.ValueOf(a)

	switch {
	case !v.IsValid():
		return Objectified{err: errors.New("invalid value")}
	case v.Kind() != reflect.Ptr:
		panic("expected a pointer")
	case v.IsNil():
		panic("cannot objectify nil pointer")
	}

	v = v.Elem()

	var o value
	var err error

	switch v.Kind() {
	case reflect.Ptr:
		o = ptrValue{Value: v, root: true}
	case reflect.Struct:
		o = structValue{Value: v, root: true}
	default:
		o, err = objectify(v)
		if err != nil {
			return Objectified{err: err}
		}
	}

	return Objectified{value: o}
}

func (o Objectified) ObjectType() string {
	return o.value.ObjectType()
}

func (o Objectified) WriteTo(w io.Writer) (n int64, err error) {
	return o.value.WriteTo(w)
}

func (o Objectified) ReadFrom(r io.Reader) (n int64, err error) {
	return o.value.ReadFrom(r)
}

func (o Objectified) UnmarshalJSON(bytes []byte) error {
	return o.value.UnmarshalJSON(bytes)
}

func (o Objectified) MarshalJSON() ([]byte, error) {
	return o.value.MarshalJSON()
}

// value is an interface for values that can be converted to an Object.
type value interface {
	Object
	json.Marshaler
	json.Unmarshaler
}

func objectify(v reflect.Value) (value, error) {
	if v.Kind() != reflect.Interface {
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
