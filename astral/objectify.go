package astral

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
)

// presenceFlagOne is the constant byte written before each value-kind element of a slice,
// array, or map value-half so []T and []*T described by the same SliceSpec produce identical
// bytes. ptrValue emits its own nil-flag and interfaceValue emits a type tag — those kinds
// don't get this byte.
var presenceFlagOne = []byte{1}

// elemNeedsPresenceFlag reports whether the container codec must synthesise the presence byte
// for elements of t. False for Ptr (ptrValue handles framing) and Interface (interfaceValue
// handles framing); true for every other kind, which writes its payload bare.
func elemNeedsPresenceFlag(t reflect.Type) bool {
	k := t.Kind()
	return k != reflect.Ptr && k != reflect.Interface
}

// consumePresenceFlag reads the synthesised presence byte. Only 0x01 is valid — value-typed
// slots have no notion of absent.
func consumePresenceFlag(r io.Reader) (int64, error) {
	var flag uint8
	if err := binary.Read(r, ByteOrder, &flag); err != nil {
		return 0, err
	}
	if flag != 1 {
		return 1, fmt.Errorf("invalid presence flag %d", flag)
	}
	return 1, nil
}

// Objectified is an Object view over a pointer to an arbitrary Go value, dispatching encode/decode
// through reflection. Construct it with Objectify; the wrapped Value must be an addressable pointer.
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

	case reflect.Uint64:
		return uint64Value{v}, nil

	case reflect.Int8:
		return int8Value{v}, nil

	case reflect.Int16:
		return int16Value{v}, nil

	case reflect.Int32:
		return int32Value{v}, nil

	case reflect.Int64:
		return int64Value{v}, nil

	// why: platform-width int/uint are rejected for the same reason supportedMapKey rejects
	// reflect.Uint — their width is platform-dependent, so silently aliasing them to int64/uint64
	// would let a struct compiled on a 32-bit host content-hash differently than the 64-bit one
	// if the codec ever started using the actual platform width. Force callers to declare the
	// width explicitly (int64/uint64) so Blueprint derivation can describe every encodable struct.
	case reflect.Int, reflect.Uint:
		return nil, fmt.Errorf("unsupported type %s %s: platform-width int/uint not allowed, use int64/uint64", v.Kind(), v.Type())

	case reflect.Array:
		return arrayValue{v}, nil

	case reflect.Slice:
		return sliceValue{v}, nil

	case reflect.String:
		return stringValue{v}, nil

	case reflect.Map:
		if _, ok := supportedMapKey(v.Type().Key().Kind()); !ok {
			return nil, fmt.Errorf("unsupported map key kind %s in %s", v.Type().Key().Kind(), v.Type())
		}
		return mapValue{v}, nil

	case reflect.Struct:
		return structValue{v, false}, nil

	case reflect.Interface:
		return interfaceValue{v}, nil

	case reflect.Bool:
		return boolValue{v}, nil

	case reflect.Float32:
		return float32Value{v}, nil

	case reflect.Float64:
		return float64Value{v}, nil

	default:
		return nil, errors.New("unsupported type " + v.Kind().String() + " " + v.Type().String())
	}
}
