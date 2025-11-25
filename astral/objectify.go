package astral

import (
	"errors"
	"reflect"
)

// Objectify converts any value into an Object. If the given value is a struct with an ObjectType() method,
// that method will be used as the returned object's type; otherwise the type is empty.
func Objectify(a any) (Object, error) {
	var v = reflect.ValueOf(a)
	if !v.IsValid() {
		return nil, errors.New("invalid value")
	}

	if v.Kind() == reflect.Interface {
		return nil, errors.New("cannot objectify an interface")
	}

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	return objectify(v)
}

func objectify(v reflect.Value) (Object, error) {
	switch v.Kind() {
	case reflect.Ptr:
		return &ptrValue{v}, nil

	case reflect.Uint8:
		return &uint8Value{v}, nil

	case reflect.Uint16:
		return &uint16Value{v}, nil

	case reflect.Uint32:
		return &uint32Value{v}, nil

	case reflect.Uint64, reflect.Uint:
		return &uint64Value{v}, nil

	case reflect.Int8:
		return &int8Value{v}, nil

	case reflect.Int16:
		return &int16Value{v}, nil

	case reflect.Int32:
		return &int32Value{v}, nil

	case reflect.Int64, reflect.Int:
		return &int64Value{v}, nil

	case reflect.Array:
		return arrayValue{v}, nil

	case reflect.Slice:
		return sliceValue{v}, nil

	case reflect.String:
		return stringValue{v}, nil

	case reflect.Map:
		return mapValue{v}, nil

	case reflect.Struct:
		return structValue{v}, nil

	case reflect.Interface:
		return interfaceValue{v}, nil

	case reflect.Bool:
		return boolValue{v}, nil

	default:
		return nil, errors.New("unsupported type " + v.Kind().String() + " " + v.Type().String())
	}
}
