package astral

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"reflect"
)

type ptrValue struct {
	reflect.Value
	skipNilFlag bool
}

var _ Object = &ptrValue{}

func (p ptrValue) ObjectType() string {
	if p.IsNil() {
		return ""
	}

	o, err := objectify(p.Elem())
	if err != nil {
		return ""
	}

	return o.ObjectType()
}

func (p ptrValue) WriteTo(w io.Writer) (n int64, err error) {
	if p.IsNil() {
		if p.skipNilFlag {
			return 0, nil
		}

		err = binary.Write(w, encoding, uint8(0)) // nil flag
		if err == nil {
			return 1, nil
		}
		return
	}

	var o Object
	o, err = objectify(p.Elem())
	if err != nil {
		return 0, err
	}

	if !p.skipNilFlag {
		err = binary.Write(w, encoding, uint8(1)) // nil flag
		if err != nil {
			return 0, err
		}
		n += 1
	}

	var m int64
	m, err = o.WriteTo(w)
	n += m
	return n, err
}

func (p ptrValue) ReadFrom(r io.Reader) (n int64, err error) {
	if !p.skipNilFlag {
		var nilFlag uint8
		err = binary.Read(r, encoding, &nilFlag)
		if err != nil {
			return
		}
		n += 1

		switch nilFlag {
		case 0:
			if p.CanSet() {
				p.Set(reflect.Zero(p.Type()))
			}
			return 1, nil
		case 1:
		default:
			return 1, errors.New("invalid nil flag")
		}
	}

	// initialize the element
	if p.CanSet() {
		p.Set(reflect.New(p.Type().Elem()))
	}

	var o Object
	o, err = objectify(p.Elem())
	if err != nil {
		return 1, err
	}

	var m int64
	m, err = o.ReadFrom(r)
	n += m
	return n, err
}

func (p ptrValue) MarshalJSON() ([]byte, error) {
	if p.IsNil() {
		return json.Marshal(nil)
	}

	e, err := objectify(p.Elem())
	if err != nil {
		return nil, err
	}

	return e.MarshalJSON()
}

func (p ptrValue) UnmarshalJSON(i []byte) error {
	if !p.CanSet() {
		return errors.New("cannot set pointer value")
	}

	if bytes.Equal(i, jsonNull) {
		p.Set(reflect.Zero(p.Type()))
		return nil
	}

	p.Set(reflect.New(p.Type().Elem()))

	o, err := objectify(p.Elem())
	if err != nil {
		return err
	}

	return o.UnmarshalJSON(i)
}
