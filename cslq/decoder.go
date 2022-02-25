package cslq

import (
	"encoding/binary"
	"errors"
	"io"
	"reflect"
	"strings"
)

// Decoder represents a CSLQ decoder.
type Decoder struct {
	r io.Reader
}

// Unmarshaler is an interface implemented by objects that can decode CSLQ representation of themselves.
type Unmarshaler interface {
	UnmarshalCSLQ(dec *Decoder) error
}

// NewDecoder returns a new Decoder instance that reads from the provided io.Reader.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// Decode decodes vars using the pattern.
func Decode(r io.Reader, pattern string, vars ...interface{}) error {
	return NewDecoder(r).Decode(pattern, vars...)
}

// Decode decodes vars using the pattern.
func (r *Decoder) Decode(pattern string, vars ...interface{}) error {
	s, err := CompileString(pattern)
	if err != nil {
		return err
	}

	return r.decode(s, vars...)
}

func (r *Decoder) decode(op OpStruct, vars ...interface{}) error {
	if len(op) != len(vars) {
		return ErrInvalidDataLength{len(op), len(vars)}
	}

	for i := range op {
		if err := r.decodeVar(op[i], vars[i]); err != nil {
			return err
		}
	}
	return nil
}

func (r *Decoder) decodeVar(op interface{}, v interface{}) error {
	var rv = reflect.ValueOf(v)

	if rv.Kind() != reflect.Ptr {
		return ErrNotAPointer
	}

	rv = rv.Elem()

	if !rv.CanSet() {
		return errors.New("cannot set variable")
	}

	var buf reflect.Value

	switch typedOp := op.(type) {
	case OpUint8:
		var i uint8
		buf = reflect.ValueOf(&i)
	case OpUint16:
		var i uint16
		buf = reflect.ValueOf(&i)
	case OpUint32:
		var i uint32
		buf = reflect.ValueOf(&i)
	case OpUint64:
		var i uint64
		buf = reflect.ValueOf(&i)
	case OpArray:
		return r.decodeArray(typedOp, v)
	case OpStruct:
		return r.decodeStruct(typedOp, v)
	case OpInterface:
		return r.decodeInterface(v)
	default:
		return ErrInvalidOp{op}
	}

	if err := binary.Read(r.r, byteOrder, buf.Interface()); err != nil {
		return err
	}

	if rv.Kind() == reflect.Bool {
		rv.SetBool(!buf.Elem().IsZero())
	} else {
		rv.Set(buf.Elem().Convert(rv.Type()))
	}

	return nil
}

func (r *Decoder) decodeArray(op OpArray, v interface{}) error {
	var length int

	if op.LenOp == nil {
		length = op.FixedLen
	} else {
		if err := r.decodeVar(op.LenOp, &length); err != nil {
			return err
		}
	}

	var rv = reflect.ValueOf(v)

	if rv.Kind() != reflect.Ptr {
		return ErrNotAPointer
	}

	rv = rv.Elem()

	if rv.Kind() == reflect.String {
		if _, ok := op.ElemOp.(OpUint8); !ok {
			return ErrCannotDecodeString
		}

		var buf = make([]byte, length)
		_, err := io.ReadFull(r.r, buf)
		if err != nil {
			return err
		}

		rv.SetString(string(buf))

		return nil
	}

	if (rv.Kind() != reflect.Array) && rv.IsNil() {
		rv.Set(reflect.MakeSlice(rv.Type(), length, length))
	}

	if rv.Len() != length {
		return ErrInvalidDataLength{length, rv.Len()}
	}

	for i := 0; i < length; i++ {
		err := r.decodeVar(op.ElemOp, rv.Index(i).Addr().Interface())
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Decoder) decodeStruct(op OpStruct, v interface{}) error {
	var rv = reflect.ValueOf(v)

	if rv.Kind() != reflect.Ptr {
		return ErrNotAPointer
	}

	rv = rv.Elem()

	if rv.Kind() != reflect.Struct {
		return ErrNotAStructPointer
	}

	vars := make([]interface{}, 0)
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		if !field.CanInterface() {
			continue
		}
		if strings.Contains(rv.Type().Field(i).Tag.Get(tagCSLQ), tagSkip) {
			continue
		}
		vars = append(vars, field.Addr().Interface())
	}

	return r.decode(op, vars...)
}

func (r *Decoder) decodeInterface(v interface{}) error {
	u, ok := v.(Unmarshaler)
	if !ok {
		return errors.New("variable does not implement Unmarshaler interface")
	}

	return u.UnmarshalCSLQ(r)
}
