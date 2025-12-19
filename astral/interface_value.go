package astral

import (
	"encoding/binary"
	"encoding/json"
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

func (i interfaceValue) IsElemNilPtr() bool {
	return i.Elem().Kind() == reflect.Ptr && i.Elem().IsNil()
}

func (i interfaceValue) MarshalJSON() ([]byte, error) {
	if !i.IsValid() || i.IsNil() {
		return json.Marshal(nil)
	}

	var j JSONEncodeAdapter

	if raw, ok := i.Interface().(*RawObject); ok {
		return json.Marshal(JSONEncodeAdapter{
			Type:    raw.ObjectType(),
			Payload: raw.Payload,
		})
	}

	switch i.Elem().Kind() {
	case reflect.Struct:
		j = JSONEncodeAdapter{}

	default:
		j = JSONEncodeAdapter{
			Object: i.Interface(),
		}
	}

	return json.Marshal(j)
}

func (i interfaceValue) UnmarshalJSON(data []byte) error {
	//TODO implement me
	panic("implement me")
}
