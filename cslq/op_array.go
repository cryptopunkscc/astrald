package cslq

import (
	"io"
	"reflect"
)

// OpArray represents an array of same type elements
type OpArray struct {
	FixedLen int
	LenOp    Op
	ElemOp   Op
}

func (op OpArray) Encode(w io.Writer, data *Fifo) error {
	var (
		length int
		rv     = reflect.ValueOf(data.Pop())
	)

	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	// encode len if necessary
	if op.FixedLen > 0 {
		length = op.FixedLen
		if rv.Len() != length {
			return ErrInvalidDataLength{op.FixedLen, rv.Len()}
		}
	} else {
		length = rv.Len()
		if err := op.LenOp.Encode(w, NewFifo(length)); err != nil {
			return err
		}
	}

	// if it's a string or []byte write everything with a single Write
	if _, ok := op.ElemOp.(OpUint8); ok {
		switch typed := rv.Interface().(type) {
		case string:
			_, err := w.Write([]byte(typed))
			return err
		case []byte:
			_, err := w.Write(typed)
			return err
		}
	}

	// encode each element
	for i := 0; i < length; i++ {
		err := op.ElemOp.Encode(w, NewFifo(rv.Index(i).Interface()))
		if err != nil {
			return err
		}
	}

	return nil
}

func (op OpArray) Decode(r io.Reader, data *Fifo) error {
	var length int

	if op.FixedLen > 0 {
		length = op.FixedLen
	} else {
		if err := op.LenOp.Decode(r, NewFifo(&length)); err != nil {
			return err
		}
	}

	var rv = reflect.ValueOf(data.Pop())

	if rv.Kind() != reflect.Ptr {
		return ErrNotAPointer
	}

	rv = rv.Elem()

	if rv.Kind() == reflect.String {
		if _, ok := op.ElemOp.(OpUint8); !ok {
			return ErrCannotDecodeString
		}

		var buf = make([]byte, length)
		_, err := io.ReadFull(r, buf)
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
		err := op.ElemOp.Decode(r, NewFifo(rv.Index(i).Addr().Interface()))
		if err != nil {
			return err
		}
	}

	return nil
}
