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

// astral:blueprint-ignore
func (i interfaceValue) ObjectType() string {
	return ""
}

func (i interfaceValue) WriteTo(w io.Writer) (n int64, err error) {
	if i.IsNil() || i.IsElemNilPtr() {
		err = binary.Write(w, ByteOrder, uint8(0)) // zero-length type means nil
		if err == nil {
			return 1, nil
		}
		return
	}

	var objectType string
	var objectWriter io.WriterTo

	switch i.Elem().Kind() {
	case reflect.Ptr:
		v := ptrValue{Value: i.Elem(), skipNilFlag: true}
		objectWriter = v
		objectType = v.ObjectType()

	case reflect.String, reflect.Slice:
		// this is a special case needed to handle various String* and Bytes* alias types
		ow, wok := i.Elem().Interface().(io.WriterTo)
		ot, tok := i.Elem().Interface().(ObjectTyper)
		if wok && tok {
			objectWriter = ow
			objectType = ot.ObjectType()
			break
		}

		fallthrough
	default:
		o, err := objectify(i.Elem())
		if err != nil {
			return n, err
		}

		objectType = o.ObjectType()
		objectWriter = o
	}

	if objectType == "" {
		return n, errors.New("interface contains an untyped object")
	}

	var m int64
	m, err = String8(objectType).WriteTo(w)
	n += m
	if err != nil {
		return
	}

	m, err = objectWriter.WriteTo(w)
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

	o := ExtractBlueprints(r).New(objectType)
	if o == nil {
		return n, newErrBlueprintNotFound(objectType)
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

	if _, ok := i.Interface().(*UnparsedObject); ok {
		return nil, errors.New("interface contains an unparsed object")
	}

	o, err := objectify(i.Elem())
	if err != nil {
		return nil, err
	}

	if o.ObjectType() == "" {
		return nil, errors.New("object behind interface has no type")
	}

	jsonBytes, err := o.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return json.Marshal(JSONAdapter{
		Type:   o.ObjectType(),
		Object: jsonBytes,
	})
}

func (i interfaceValue) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, jsonNull) {
		i.SetZero()
		return nil
	}

	var j JSONAdapter
	err := json.Unmarshal(data, &j)
	if err != nil {
		return err
	}

	o := DefaultBlueprints.New(j.Type)
	if o == nil {
		return newErrBlueprintNotFound(j.Type)
	}

	if j.Object != nil {
		err = json.Unmarshal(j.Object, o)
		if err != nil {
			return err
		}
	}

	i.Set(reflect.ValueOf(o).Convert(i.Type()))

	return nil
}

func (i interfaceValue) IsElemNilPtr() bool {
	return i.Elem().Kind() == reflect.Ptr && i.Elem().IsNil()
}
