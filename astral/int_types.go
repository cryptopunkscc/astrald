package astral

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"strconv"
)

type Int8 int8

func NewInt8(i int8) *Int8 {
	return (*Int8)(&i)
}

// astral:blueprint-ignore
func (Int8) ObjectType() string {
	return "int8"
}

func (i Int8) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, ByteOrder, i)
	if err == nil {
		n = 1
	}
	return
}

func (i *Int8) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, ByteOrder, i)
	if err == nil {
		n = 1
	}
	return
}

func (i Int8) MarshalJSON() ([]byte, error) {
	return json.Marshal(int8(i))
}

func (i *Int8) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, (*int8)(i))
}

func (i Int8) MarshalText() ([]byte, error) {
	return []byte(strconv.FormatInt(int64(i), 10)), nil
}

func (i *Int8) UnmarshalText(text []byte) error {
	v, err := strconv.ParseInt(string(text), 10, 8)
	if err != nil {
		return err
	}
	*i = Int8(v)
	return nil
}

func (i Int8) String() string {
	return strconv.FormatInt(int64(i), 10)
}

type Int16 int16

func NewInt16(i int16) *Int16 {
	return (*Int16)(&i)
}

// astral:blueprint-ignore
func (Int16) ObjectType() string {
	return "int16"
}

func (i Int16) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, ByteOrder, i)
	if err == nil {
		n = 1
	}
	return
}

func (i *Int16) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, ByteOrder, i)
	if err == nil {
		n = 1
	}
	return
}

func (i Int16) MarshalJSON() ([]byte, error) {
	return json.Marshal(int16(i))
}

func (i *Int16) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, (*int16)(i))
}

func (i Int16) MarshalText() ([]byte, error) {
	return []byte(strconv.FormatInt(int64(i), 10)), nil
}

func (i *Int16) UnmarshalText(text []byte) error {
	v, err := strconv.ParseInt(string(text), 10, 8)
	if err != nil {
		return err
	}
	*i = Int16(v)
	return nil
}

func (i Int16) String() string {
	return strconv.FormatInt(int64(i), 10)
}

type Int32 int32

func NewInt32(i int32) *Int32 {
	return (*Int32)(&i)
}

// astral:blueprint-ignore
func (Int32) ObjectType() string {
	return "int32"
}

func (i Int32) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, ByteOrder, i)
	if err == nil {
		n = 1
	}
	return
}

func (i *Int32) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, ByteOrder, i)
	if err == nil {
		n = 1
	}
	return
}

func (i Int32) MarshalJSON() ([]byte, error) {
	return json.Marshal(int32(i))
}

func (i *Int32) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, (*int32)(i))
}

func (i Int32) MarshalText() ([]byte, error) {
	return []byte(strconv.FormatInt(int64(i), 10)), nil
}

func (i *Int32) UnmarshalText(text []byte) error {
	v, err := strconv.ParseInt(string(text), 10, 8)
	if err != nil {
		return err
	}
	*i = Int32(v)
	return nil
}

func (i Int32) String() string {
	return strconv.FormatInt(int64(i), 10)
}

type Int64 int64

func NewInt64(i int64) *Int64 {
	return (*Int64)(&i)
}

// astral:blueprint-ignore
func (Int64) ObjectType() string {
	return "int64"
}

func (i Int64) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, ByteOrder, i)
	if err == nil {
		n = 1
	}
	return
}

func (i *Int64) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, ByteOrder, i)
	if err == nil {
		n = 1
	}
	return
}

func (i Int64) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(i))
}

func (i *Int64) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, (*int64)(i))
}

func (i Int64) MarshalText() ([]byte, error) {
	return []byte(strconv.FormatInt(int64(i), 10)), nil
}

func (i *Int64) UnmarshalText(text []byte) error {
	v, err := strconv.ParseInt(string(text), 10, 8)
	if err != nil {
		return err
	}
	*i = Int64(v)
	return nil
}

func (i Int64) String() string {
	return strconv.FormatInt(int64(i), 10)
}

func init() {
	var (
		i8  Int8
		i16 Int16
		i32 Int32
		i64 Int64
	)
	_ = Add(&i8, &i16, &i32, &i64)
}
