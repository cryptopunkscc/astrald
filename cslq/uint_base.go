package cslq

import (
	"encoding/binary"
	"io"
	"reflect"
)

type uintBase struct{}

func (op uintBase) Encode(w io.Writer, data *Fifo, targetType reflect.Type) error {
	var rv = reflect.ValueOf(data.Pop())

	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Bool {
		if rv.Bool() {
			return binary.Write(w, byteOrder, reflect.ValueOf(1).Convert(targetType).Interface())
		} else {
			return binary.Write(w, byteOrder, reflect.ValueOf(0).Convert(targetType).Interface())
		}
	}

	if !rv.CanConvert(targetType) {
		return ErrCannotConvert{rv.Type().String(), targetType.String()}
	}

	return binary.Write(w, byteOrder, rv.Convert(targetType).Interface())
}

func (op uintBase) Decode(r io.Reader, data *Fifo, i interface{}) error {
	if err := binary.Read(r, byteOrder, i); err != nil {
		return err
	}

	var rv = reflect.ValueOf(data.Pop())

	if rv.Kind() != reflect.Ptr {
		return ErrNotAPointer
	}
	rv = rv.Elem()

	switch rv.Kind() {
	case reflect.Bool:
		rv.SetBool(reflect.ValueOf(i).Elem().Uint() != 0)
	default:
		rv.Set(reflect.ValueOf(i).Elem().Convert(rv.Type()))
	}

	return nil
}
