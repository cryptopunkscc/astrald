package astral

import (
	"encoding/binary"
	"errors"
	"io"
	"reflect"
	"strconv"
)

type uint8Value struct {
	reflect.Value
}

type uint16Value struct {
	reflect.Value
}

type uint32Value struct {
	reflect.Value
}

type uint64Value struct {
	reflect.Value
}

type int8Value struct {
	reflect.Value
}

type int16Value struct {
	reflect.Value
}

type int32Value struct {
	reflect.Value
}

type int64Value struct {
	reflect.Value
}

var _ Object = &uint8Value{}
var _ Object = &int8Value{}
var _ Object = &int16Value{}
var _ Object = &int32Value{}
var _ Object = &int64Value{}
var _ Object = &uint16Value{}
var _ Object = &uint32Value{}
var _ Object = &uint64Value{}

func (u uint8Value) ObjectType() string {
	return "uint8"
}

func (u uint8Value) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, encoding, uint8(u.Uint()))
	if err == nil {
		n = 1
	}
	return
}

func (u uint8Value) ReadFrom(r io.Reader) (n int64, err error) {
	if !u.CanSet() {
		return 0, errors.New("cannot set value")
	}

	var i uint8
	err = binary.Read(r, encoding, &i)

	u.SetUint(uint64(i))

	return 1, nil
}

func (u uint8Value) String() string {
	return strconv.FormatUint(u.Uint(), 10)
}

func (i int8Value) ObjectType() string {
	return "int8"
}

func (i int8Value) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, encoding, int8(i.Int()))
	if err == nil {
		n = 1
	}
	return
}

func (i int8Value) ReadFrom(r io.Reader) (n int64, err error) {
	if !i.CanSet() {
		return 0, errors.New("cannot set value")
	}

	var v int8
	err = binary.Read(r, encoding, &v)

	i.SetInt(int64(v))

	return 1, nil
}

func (i int8Value) String() string {
	return strconv.FormatInt(i.Int(), 10)
}

func (i int16Value) ObjectType() string {
	return "int16"
}

func (i int16Value) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, encoding, int16(i.Int()))
	if err == nil {
		n = 2
	}
	return
}

func (i int16Value) ReadFrom(r io.Reader) (n int64, err error) {
	if !i.CanSet() {
		return 0, errors.New("cannot set value")
	}

	var v int16
	err = binary.Read(r, encoding, &v)

	i.SetInt(int64(v))

	return 2, nil
}

func (i int16Value) String() string {
	return strconv.FormatInt(i.Int(), 10)
}

func (i int32Value) ObjectType() string {
	return "int32"
}

func (i int32Value) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, encoding, int32(i.Int()))
	if err == nil {
		n = 4
	}
	return
}

func (i int32Value) ReadFrom(r io.Reader) (n int64, err error) {
	if !i.CanSet() {
		return 0, errors.New("cannot set value")
	}

	var v int32
	err = binary.Read(r, encoding, &v)

	i.SetInt(int64(v))

	return 4, nil
}

func (i int32Value) String() string {
	return strconv.FormatInt(i.Int(), 10)
}

func (i int64Value) ObjectType() string {
	return "int64"
}

func (i int64Value) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, encoding, i.Int())
	if err == nil {
		n = 8
	}
	return
}

func (i int64Value) ReadFrom(r io.Reader) (n int64, err error) {
	if !i.CanSet() {
		return 0, errors.New("cannot set value")
	}

	var v int64
	err = binary.Read(r, encoding, &v)

	i.SetInt(v)

	return 8, nil
}

func (i int64Value) String() string {
	return strconv.FormatInt(i.Int(), 10)
}

func (u uint16Value) ObjectType() string {
	return "uint16"
}

func (u uint16Value) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, encoding, uint16(u.Uint()))
	if err == nil {
		n = 2
	}
	return
}

func (u uint16Value) ReadFrom(r io.Reader) (n int64, err error) {
	if !u.CanSet() {
		return 0, errors.New("cannot set value")
	}

	var v uint16
	err = binary.Read(r, encoding, &v)

	u.SetUint(uint64(v))

	return 2, nil
}

func (u uint16Value) String() string {
	return strconv.FormatUint(u.Uint(), 10)
}

func (u uint32Value) ObjectType() string {
	return "uint32"
}

func (u uint32Value) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, encoding, uint32(u.Uint()))
	if err == nil {
		n = 4
	}
	return
}

func (u uint32Value) ReadFrom(r io.Reader) (n int64, err error) {
	if !u.CanSet() {
		return 0, errors.New("cannot set value")
	}

	var v uint32
	err = binary.Read(r, encoding, &v)

	u.SetUint(uint64(v))

	return 4, nil
}

func (u uint32Value) String() string {
	return strconv.FormatUint(u.Uint(), 10)
}

func (u uint64Value) ObjectType() string {
	return "uint64"
}

func (u uint64Value) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, encoding, u.Uint())
	if err == nil {
		n = 8
	}
	return
}

func (u uint64Value) ReadFrom(r io.Reader) (n int64, err error) {
	if !u.CanSet() {
		return 0, errors.New("cannot set value")
	}

	var v uint64
	err = binary.Read(r, encoding, &v)

	u.SetUint(v)

	return 8, nil
}

func (u uint64Value) String() string {
	return strconv.FormatUint(u.Uint(), 10)
}
