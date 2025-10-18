package astral

import (
	"encoding/binary"
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

// String methods
func (u Uint8) String() string {
	return strconv.FormatUint(uint64(u), 10)
}
func (u Uint16) String() string {
	return strconv.FormatUint(uint64(u), 10)
}
func (u Uint32) String() string {
	return strconv.FormatUint(uint64(u), 10)
}
func (u Uint64) String() string {
	return strconv.FormatUint(uint64(u), 10)
}
func (i Int8) String() string {
	return strconv.FormatInt(int64(i), 10)
}
func (i Int16) String() string {
	return strconv.FormatInt(int64(i), 10)
}
func (i Int32) String() string {
	return strconv.FormatInt(int64(i), 10)
}
func (i Int64) String() string {
	return strconv.FormatInt(int64(i), 10)
}

// MarshalText methods
func (u Uint8) MarshalText() ([]byte, error) {
	return []byte(u.String()), nil
}
func (u Uint16) MarshalText() ([]byte, error) {
	return []byte(u.String()), nil
}
func (u Uint32) MarshalText() ([]byte, error) {
	return []byte(u.String()), nil
}
func (u Uint64) MarshalText() ([]byte, error) {
	return []byte(u.String()), nil
}
func (i Int8) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}
func (i Int16) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}
func (i Int32) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}
func (i Int64) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

// UnmarshalText methods
func (u *Uint8) UnmarshalText(text []byte) error {
	v, err := strconv.ParseUint(string(text), 10, 8)
	if err != nil {
		return err
	}
	*u = Uint8(v)
	return nil
}
func (u *Uint16) UnmarshalText(text []byte) error {
	v, err := strconv.ParseUint(string(text), 10, 16)
	if err != nil {
		return err
	}
	*u = Uint16(v)
	return nil
}
func (u *Uint32) UnmarshalText(text []byte) error {
	v, err := strconv.ParseUint(string(text), 10, 32)
	if err != nil {
		return err
	}
	*u = Uint32(v)
	return nil
}
func (u *Uint64) UnmarshalText(text []byte) error {
	v, err := strconv.ParseUint(string(text), 10, 64)
	if err != nil {
		return err
	}
	*u = Uint64(v)
	return nil
}
func (i *Int8) UnmarshalText(text []byte) error {
	v, err := strconv.ParseInt(string(text), 10, 8)
	if err != nil {
		return err
	}
	*i = Int8(v)
	return nil
}
func (i *Int16) UnmarshalText(text []byte) error {
	v, err := strconv.ParseInt(string(text), 10, 16)
	if err != nil {
		return err
	}
	*i = Int16(v)
	return nil
}
func (i *Int32) UnmarshalText(text []byte) error {
	v, err := strconv.ParseInt(string(text), 10, 32)
	if err != nil {
		return err
	}
	*i = Int32(v)
	return nil
}
func (i *Int64) UnmarshalText(text []byte) error {
	v, err := strconv.ParseInt(string(text), 10, 64)
	if err != nil {
		return err
	}
	*i = Int64(v)
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
