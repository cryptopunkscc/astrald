package astral

import (
	"encoding/binary"
	"errors"
	"io"
	"reflect"
	"strconv"
)

type float32Value struct {
	reflect.Value
}

var _ Object = &float32Value{}

// astral:blueprint-ignore
func (f float32Value) ObjectType() string {
	return "float32"
}

func (f float32Value) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, ByteOrder, float32(f.Float()))
	if err == nil {
		n = 4
	}
	return
}

func (f float32Value) ReadFrom(r io.Reader) (n int64, err error) {
	if !f.CanSet() {
		return 0, errors.New("cannot set value")
	}

	var v float32
	err = binary.Read(r, ByteOrder, &v)

	f.SetFloat(float64(v))

	return 4, nil
}

func (f float32Value) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatFloat(f.Float(), 'f', -1, 32)), nil
}

func (f float32Value) UnmarshalJSON(bytes []byte) error {
	v, err := strconv.ParseFloat(string(bytes), 32)
	if err != nil {
		return err
	}
	f.SetFloat(v)
	return nil
}

func (f float32Value) String() string {
	return strconv.FormatFloat(f.Float(), 'f', -1, 32)
}

type float64Value struct {
	reflect.Value
}

var _ Object = &float64Value{}

// astral:blueprint-ignore
func (f float64Value) ObjectType() string {
	return "float64"
}

func (f float64Value) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, ByteOrder, f.Float())
	if err == nil {
		n = 8
	}
	return
}

func (f float64Value) ReadFrom(r io.Reader) (n int64, err error) {
	if !f.CanSet() {
		return 0, errors.New("cannot set value")
	}

	var v float64
	err = binary.Read(r, ByteOrder, &v)

	f.SetFloat(v)

	return 8, nil
}

func (f float64Value) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatFloat(f.Float(), 'f', -1, 64)), nil
}

func (f float64Value) UnmarshalJSON(bytes []byte) error {
	v, err := strconv.ParseFloat(string(bytes), 64)
	if err != nil {
		return err
	}
	f.SetFloat(v)
	return nil
}

func (f float64Value) String() string {
	return strconv.FormatFloat(f.Float(), 'f', -1, 64)
}
