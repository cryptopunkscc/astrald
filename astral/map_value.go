package astral

import (
	"bytes"
	encoding2 "encoding"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

type mapValue struct {
	reflect.Value
}

var _ Object = &mapValue{}

func (val mapValue) ObjectType() string {
	return ""
}

func (val mapValue) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, encoding, uint32(val.Len()))
	if err != nil {
		return
	}
	n += 4

	var o Object
	var i int64

	for _, k := range val.MapKeys() {
		nkey := k.Kind() == reflect.Ptr || k.Kind() == reflect.Interface
		if wto, ok := k.Interface().(io.WriterTo); ok && !nkey {
			i, err = wto.WriteTo(w)
		} else {
			o, err = objectify(k)
			if err != nil {
				return
			}

			i, err = o.WriteTo(w)
		}

		n += i
		if err != nil {
			return
		}

		v := val.MapIndex(k)

		nval := v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface
		if wto, ok := v.Interface().(io.WriterTo); ok && !nval {
			i, err = wto.WriteTo(w)
		} else {
			o, err = objectify(v)
			if err != nil {
				return
			}

			i, err = o.WriteTo(w)
		}

		n += i
		if err != nil {
			return
		}
	}

	return
}

func (val mapValue) ReadFrom(r io.Reader) (n int64, err error) {
	var l uint32
	err = binary.Read(r, encoding, &l)
	if err != nil {
		return
	}
	n += 4

	if l == 0 {
		val.SetZero()
		return
	}

	val.Set(reflect.MakeMap(val.Type()))

	var o Object
	var k int64

	for range l {
		var key = reflect.New(val.Type().Key()).Elem()

		o, err = objectify(key)
		k, err = o.ReadFrom(r)
		n += k
		if err != nil {
			return
		}

		var value = reflect.New(val.Type().Elem()).Elem()
		o, err = objectify(value)
		k, err = o.ReadFrom(r)
		n += k
		if err != nil {
			return
		}

		val.SetMapIndex(key, value)
	}

	return
}

func (val mapValue) MarshalJSON() ([]byte, error) {
	if val.IsNil() {
		return jsonNull, nil
	}

	var jmap = map[string]json.RawMessage{}
	var err error

	for _, mapKey := range val.MapKeys() {
		var key []byte

		m, ok := mapKey.Interface().(encoding2.TextMarshaler)
		if ok {
			key, err = m.MarshalText()
			if err != nil {
				return nil, err
			}
		} else {
			switch mapKey.Kind() {
			case reflect.String:
				key = []byte(mapKey.String())
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				key = []byte(strconv.FormatInt(mapKey.Int(), 10))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				key = []byte(strconv.FormatUint(mapKey.Uint(), 10))
			default:
				return nil, errors.New("map key does not implement text encoding")
			}
		}

		mapValue := val.MapIndex(mapKey)

		o, err := objectify(mapValue)
		if err != nil {
			return nil, err
		}

		value, err := o.MarshalJSON()
		if err != nil {
			return nil, err
		}

		jmap[string(key)] = value
	}

	return json.Marshal(jmap)
}

func (val mapValue) UnmarshalJSON(data []byte) error {
	if bytes.Compare(data, jsonNull) == 0 {
		val.SetZero()
		return nil
	}

	val.Set(reflect.MakeMap(val.Type()))

	var fields map[string]json.RawMessage
	err := json.Unmarshal(data, &fields)
	if err != nil {
		return err
	}

	keyType := val.Type().Key()

	for k, v := range fields {
		var mapVal reflect.Value
		var mapKey = reflect.New(keyType)

		if u, ok := mapKey.Interface().(encoding2.TextUnmarshaler); ok {
			err = u.UnmarshalText([]byte(k))
		} else {
			switch keyType.Kind() {
			case reflect.String:
				mapKey.Elem().SetString(k)

			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				i, err := strconv.ParseInt(k, 10, 64)
				if err != nil {
					return fmt.Errorf("error parsing int key: %w", err)
				}
				mapKey.Elem().SetInt(i)

			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				i, err := strconv.ParseUint(k, 10, 64)
				if err != nil {
					return fmt.Errorf("error parsing uint key: %w", err)
				}
				mapKey.Elem().SetUint(i)

			default:
				return errors.New("map key does not implement text decoding")
			}
		}

		mapVal = reflect.New(val.Type().Elem()).Elem()
		o, err := objectify(mapVal)
		if err != nil {
			return err
		}

		err = o.UnmarshalJSON(v)
		if err != nil {
			return err
		}

		val.SetMapIndex(mapKey.Elem(), mapVal)
	}

	return nil
}
