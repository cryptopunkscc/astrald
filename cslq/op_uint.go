package cslq

import (
	"io"
	"reflect"
)

// OpUint8 codes basic uint8 type
type OpUint8 struct{ uintBase }

func (op OpUint8) Encode(w io.Writer, data *Fifo) error {
	return op.uintBase.Encode(w, data, reflect.TypeOf(uint8(0)))
}

func (op OpUint8) Decode(r io.Reader, data *Fifo) error {
	var i uint8
	return op.uintBase.Decode(r, data, &i)
}

// OpUint16 codes basic uint16 type
type OpUint16 struct{ uintBase }

func (op OpUint16) Encode(w io.Writer, data *Fifo) error {
	return op.uintBase.Encode(w, data, reflect.TypeOf(uint16(0)))

}

func (op OpUint16) Decode(r io.Reader, data *Fifo) error {
	var i uint16
	return op.uintBase.Decode(r, data, &i)
}

// OpUint32 codes basic uint32 type
type OpUint32 struct{ uintBase }

func (op OpUint32) Encode(w io.Writer, data *Fifo) error {
	return op.uintBase.Encode(w, data, reflect.TypeOf(uint32(0)))
}

func (op OpUint32) Decode(r io.Reader, data *Fifo) error {
	var i uint32
	return op.uintBase.Decode(r, data, &i)
}

// OpUint64 codes basic uint64 type
type OpUint64 struct{ uintBase }

func (op OpUint64) Encode(w io.Writer, data *Fifo) error {
	return op.uintBase.Encode(w, data, reflect.TypeOf(uint64(0)))
}

func (op OpUint64) Decode(r io.Reader, data *Fifo) error {
	var i uint64
	return op.uintBase.Decode(r, data, &i)
}
