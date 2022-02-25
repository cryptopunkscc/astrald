package cslq

import (
	"encoding/binary"
	"errors"
	"io"
	"reflect"
	"strings"
)

// Encoder represents a CSLQ encoder.
type Encoder struct {
	w io.Writer
}

// Marshaler is an interface implemented by objects that can encode CSLQ representation of themselves.
type Marshaler interface {
	MarshalCSLQ(enc *Encoder) error
}

// NewEncoder returns a new Encoder instance that writes to the provided io.Writer.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// Encode encodes vars according to the pattern.
func Encode(w io.Writer, pattern string, vars ...interface{}) error {
	return NewEncoder(w).Encode(pattern, vars...)
}

// Encode encodes vars according to the pattern.
func (w *Encoder) Encode(pattern string, vars ...interface{}) error {
	s, err := CompileString(pattern)
	if err != nil {
		return err
	}

	return w.encode(s, vars...)
}

func (w *Encoder) encode(op OpStruct, vars ...interface{}) error {
	if len(op) != len(vars) {
		return ErrInvalidDataLength{len(op), len(vars)}
	}

	for i := range op {
		if err := w.encodeVar(op[i], vars[i]); err != nil {
			return err
		}
	}
	return nil
}

func (w *Encoder) encodeVar(op interface{}, v interface{}) error {
	var rv = reflect.ValueOf(v)
	var targetType reflect.Type

	switch typedOp := op.(type) {
	case OpUint8:
		targetType = reflect.TypeOf(uint8(0))
	case OpUint16:
		targetType = reflect.TypeOf(uint16(0))
	case OpUint32:
		targetType = reflect.TypeOf(uint32(0))
	case OpUint64:
		targetType = reflect.TypeOf(uint64(0))
	case OpInterface:
		return w.encodeInterface(v)
	case OpStruct:
		return w.encodeStruct(typedOp, v)
	case OpArray:
		return w.encodeArray(typedOp, v)
	default:
		return ErrInvalidOp{op}
	}

	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Bool {
		if rv.Bool() {
			rv = reflect.ValueOf(uint8(1))
		} else {
			rv = reflect.ValueOf(uint8(0))
		}
	}

	if !rv.CanConvert(targetType) {
		return ErrCannotConvert{
			rv.Type().String(),
			targetType.String(),
		}
	}
	rv = rv.Convert(targetType)
	return binary.Write(w.w, byteOrder, rv.Interface())
}

func (w *Encoder) encodeArray(op OpArray, v interface{}) error {
	var length int
	var rv = reflect.ValueOf(v)

	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if op.LenOp == nil {
		length = op.FixedLen
		if rv.Len() != length {
			return ErrInvalidDataLength{op.FixedLen, rv.Len()}
		}
	} else {
		length = rv.Len()
		if err := w.encodeVar(op.LenOp, length); err != nil {
			return err
		}
	}

	for i := 0; i < length; i++ {
		err := w.encodeVar(op.ElemOp, rv.Index(i).Interface())
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *Encoder) encodeStruct(op OpStruct, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return ErrNotAStructPointer
	}

	vars := make([]interface{}, 0)
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)

		if strings.Contains(rv.Type().Field(i).Tag.Get(tagCSLQ), tagSkip) {
			continue
		}

		if field.CanInterface() {
			vars = append(vars, field.Interface())
		}
	}

	return w.encode(op, vars...)
}

func (w *Encoder) encodeInterface(v interface{}) error {
	m, ok := v.(Marshaler)
	if !ok {
		return errors.New("variable does not implement Marshaler interface")
	}

	return m.MarshalCSLQ(w)
}
