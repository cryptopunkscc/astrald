package cslq

import (
	"io"
	"reflect"
)

// OpStruct encodes and decodes structures
type OpStruct []Op

func (op OpStruct) Encode(w io.Writer, data *Fifo) error {
	rv := reflect.ValueOf(data.Pop())
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return ErrNotAStructPointer
	}

	vars := extractStructFields(rv)

	if len(vars) != len(op) {
		return ErrInvalidDataLength{len(op), len(vars)}
	}

	for i, _op := range op {
		if err := _op.Encode(w, NewFifo(vars[i])); err != nil {
			return err
		}
	}

	return nil
}

func (op OpStruct) Decode(r io.Reader, data *Fifo) error {
	var rv = reflect.ValueOf(data.Pop())

	if rv.Kind() == reflect.Ptr && rv.Elem().Kind() == reflect.Ptr && rv.Elem().IsZero() {
		rv.Elem().Set(reflect.New(rv.Type().Elem().Elem()))
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Ptr {
		return ErrNotAPointer
	}

	rv = rv.Elem()

	if rv.Kind() != reflect.Struct {
		return ErrNotAStructPointer
	}

	vars := extractStructFields(rv)

	for i, _op := range op {
		if err := _op.Decode(r, NewFifo(vars[i])); err != nil {
			return err
		}
	}

	return nil
}

func (op OpStruct) String() string {
	var s = "{"
	for _, sub := range op {
		s = s + sub.String()
	}
	return s + "}"
}
