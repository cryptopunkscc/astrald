package astral

import (
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
)

var encoding = binary.BigEndian

type Uint8 uint8

type Uint16 uint16

type Uint32 uint32

type Uint64 uint64

type Int8 int8

type Int16 int16

type Int32 int32

type Int64 int64

func (Uint8) ObjectType() string {
	return "uint8"
}

func (Uint16) ObjectType() string {
	return "uint16"
}

func (Uint32) ObjectType() string {
	return "uint32"
}

func (Uint64) ObjectType() string {
	return "uint64"
}

func (Int8) ObjectType() string {
	return "int8"
}

func (Int16) ObjectType() string {
	return "int16"
}

func (Int32) ObjectType() string {
	return "int32"
}

func (Int64) ObjectType() string {
	return "int64"
}

func (u Uint8) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, encoding, u)
	if err == nil {
		n = 1
	}
	return
}

func (u Uint16) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, encoding, u)
	if err == nil {
		n = 2
	}
	return
}

func (u Uint32) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, encoding, u)
	if err == nil {
		n = 4
	}
	return
}

func (u Uint64) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, encoding, u)
	if err == nil {
		n = 8
	}
	return
}

func (i Int8) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, encoding, i)
	if err == nil {
		n = 1
	}
	return
}

func (i Int16) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, encoding, i)
	if err == nil {
		n = 2
	}
	return
}

func (i Int32) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, encoding, i)
	if err == nil {
		n = 4
	}
	return
}

func (i Int64) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, encoding, i)
	if err == nil {
		n = 8
	}
	return
}

func (u *Uint8) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, encoding, u)
	if err == nil {
		n = 1
	}
	return
}

func (u *Uint16) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, encoding, u)
	if err == nil {
		n = 2
	}
	return
}

func (u *Uint32) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, encoding, u)
	if err == nil {
		n = 4
	}
	return
}

func (u *Uint64) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, encoding, u)
	if err == nil {
		n = 8
	}
	return
}

func (i *Int8) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, encoding, i)
	if err == nil {
		n = 1
	}
	return
}

func (i *Int16) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, encoding, i)
	if err == nil {
		n = 2
	}
	return
}

func (i *Int32) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, encoding, i)
	if err == nil {
		n = 4
	}
	return
}

func (i *Int64) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, encoding, i)
	if err == nil {
		n = 8
	}
	return
}

func (i Uint16) MarshalText() (text []byte, err error) {
	return []byte(fmt.Sprintf(`%d`, i)), nil
}

func (i *Uint16) UnmarshalText(text []byte) error {
	i64, err := strconv.ParseInt(string(text), 10, 16)
	if err != nil {
		return err
	}

	*i = Uint16(i64)
	return nil
}

func init() {
	var (
		u8  Uint8
		u16 Uint16
		u32 Uint32
		u64 Uint64
		i8  Int8
		i16 Int16
		i32 Int32
		i64 Int64
	)
	_ = DefaultBlueprints.Add(
		&u8, &u16, &u32, &u64,
		&i8, &i16, &i32, &i64,
	)

}
