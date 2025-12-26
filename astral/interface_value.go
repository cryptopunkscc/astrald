package astral

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
)

type interfaceValue struct {
	reflect.Value
}

var _ Object = &interfaceValue{}

func (i interfaceValue) ObjectType() string {
	return ""
}

func (i interfaceValue) WriteTo(w io.Writer) (n int64, err error) {
	if i.IsNil() || i.IsElemNilPtr() {
		err = binary.Write(w, encoding, uint8(0)) // zero-length type means nil
		if err == nil {
			return 1, nil
		}
		return
	}

	var o Object

	if i.Elem().Kind() == reflect.Ptr {
		o = ptrValue{Value: i.Elem(), skipNilFlag: true}
	} else {
		o, err = objectify(i.Elem())
		if err != nil {
			return
		}
	}

	var m int64
	m, err = String8(o.ObjectType()).WriteTo(w)
	n += m
	if err != nil {
		return
	}

	m, err = o.WriteTo(w)
	n += m

	return
}

func (i interfaceValue) ReadFrom(r io.Reader) (n int64, err error) {
	var objectType string
	m, err := (*String8)(&objectType).ReadFrom(r)
	n += m
	if err != nil {
		return
	}

	if len(objectType) == 0 {
		i.Set(reflect.Zero(i.Type()))
		return
	}

	o := ExtractBlueprints(r).Make(objectType)
	if o == nil {
		return n, fmt.Errorf("unknown object type %s", objectType)
	}

	if !reflect.ValueOf(o).CanConvert(i.Type()) {
		err = fmt.Errorf("cannot convert %s to %s", reflect.TypeOf(o), i.Type())
		return
	}

	m, err = o.ReadFrom(r)
	n += m
	if err != nil {
		return
	}

	i.Set(reflect.ValueOf(o).Convert(i.Type()))

	return
}

func (i interfaceValue) MarshalJSON() ([]byte, error) {
	if !i.IsValid() || i.IsNil() || i.IsElemNilPtr() {
		return jsonNull, nil
	}

	if i.Elem().Type().Kind() == reflect.Map {
		return mapValue{Value: i.Elem()}.MarshalJSON()
	}

	var j JSONEncodeAdapter

	if raw, ok := i.Interface().(*RawObject); ok {
		return json.Marshal(JSONEncodeAdapter{
			Type:    raw.ObjectType(),
			Payload: raw.Payload,
		})
	}

	o, err := objectify(i.Elem())
	if err != nil {
		return nil, err
	}

	jdata, err := o.MarshalJSON()
	if err != nil {
		return nil, err
	}

	j = JSONEncodeAdapter{
		Type:   o.ObjectType(),
		Object: json.RawMessage(jdata),
	}

	return json.Marshal(j)
}

func (i interfaceValue) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, jsonNull) {
		i.SetZero()
		return nil
	}

	var j JSONDecodeAdapter
	err := json.Unmarshal(data, &j)
	if err != nil {
		return err
	}

	o := DefaultBlueprints.Make(j.Type)
	if o == nil {
		return errors.New("cant make object of type " + j.Type)
	}

	switch {
	case j.Object != nil:
		err = json.Unmarshal(j.Object, o)
		if err != nil {
			return err
		}
	case j.Payload != nil:
		return errors.New("payload not supported")
	}

	i.Set(reflect.ValueOf(o).Convert(i.Type()))

	return nil
}

func (i interfaceValue) IsElemNilPtr() bool {
	return i.Elem().Kind() == reflect.Ptr && i.Elem().IsNil()
}
