package astral

import (
	"bytes"
	encoding2 "encoding"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"reflect"
)

type mapValue struct {
	reflect.Value
}

var _ Object = &mapValue{}

func (m mapValue) ObjectType() string {
	return ""
}

func (m mapValue) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, encoding, uint32(m.Len()))
	if err != nil {
		return
	}
	n += 4

	var o Object
	var i int64

	for _, k := range m.MapKeys() {
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

		v := m.MapIndex(k)

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

func (m mapValue) ReadFrom(r io.Reader) (n int64, err error) {
	var l uint32
	err = binary.Read(r, encoding, &l)
	if err != nil {
		return
	}
	n += 4

	if l == 0 {
		m.SetZero()
		return
	}

	m.Set(reflect.MakeMap(m.Type()))

	var o Object
	var k int64

	for range l {
		var key = reflect.New(m.Type().Key()).Elem()

		o, err = objectify(key)
		k, err = o.ReadFrom(r)
		n += k
		if err != nil {
			return
		}

		var value = reflect.New(m.Type().Elem()).Elem()
		o, err = objectify(value)
		k, err = o.ReadFrom(r)
		n += k
		if err != nil {
			return
		}

		m.SetMapIndex(key, value)
	}

	return
}

func (m mapValue) MarshalJSON() ([]byte, error) {
	if m.IsNil() {
		return jsonNull, nil
	}

	var jmap = map[string]json.RawMessage{}

	for _, mapKey := range m.MapKeys() {
		tm, ok := mapKey.Interface().(encoding2.TextMarshaler)
		if !ok {
			return nil, errors.New("map key does not implement text encoding")
		}

		key, err := tm.MarshalText()
		if err != nil {
			return nil, err
		}

		mapValue := m.MapIndex(mapKey)

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

func (m mapValue) UnmarshalJSON(data []byte) error {
	if bytes.Compare(data, jsonNull) == 0 {
		m.SetZero()
		return nil
	}

	m.Set(reflect.MakeMap(m.Type()))

	var fields map[string]json.RawMessage
	err := json.Unmarshal(data, &fields)
	if err != nil {
		return err
	}

	keyType := m.Type().Key()

	for k, v := range fields {
		var mapKey, mapVal reflect.Value

		mapKey = reflect.New(keyType).Elem()
		o, err := objectify(mapKey)
		if err != nil {
			return err
		}

		err = o.UnmarshalJSON([]byte(k))
		if err != nil {
			return err
		}

		mapVal = reflect.New(m.Type().Elem()).Elem()
		o, err = objectify(mapVal)
		if err != nil {
			return err
		}

		err = o.UnmarshalJSON(v)
		if err != nil {
			return err
		}

		m.SetMapIndex(mapKey, mapVal)
	}

	return nil
}
