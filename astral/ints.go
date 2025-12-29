package astral

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"strconv"
)

type Uint8 uint8

func (Uint8) ObjectType() string {
	return "uint8"
}

func (u Uint8) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, ByteOrder, u)
	if err == nil {
		n = 1
	}
	return
}

func (u *Uint8) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, ByteOrder, u)
	if err == nil {
		n = 1
	}
	return
}

func (u Uint8) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint8(u))
}

func (u *Uint8) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, (*uint8)(u))
}

func (u Uint8) MarshalText() ([]byte, error) {
	return []byte(strconv.FormatUint(uint64(u), 10)), nil
}

func (u *Uint8) UnmarshalText(text []byte) error {
	v, err := strconv.ParseUint(string(text), 10, 8)
	if err != nil {
		return err
	}
	*u = Uint8(v)
	return nil
}

func (u Uint8) String() string {
	return strconv.FormatUint(uint64(u), 10)
}

type Uint16 uint16

func (Uint16) ObjectType() string {
	return "uint16"
}

func (u Uint16) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, ByteOrder, u)
	if err == nil {
		n = 1
	}
	return
}

func (u *Uint16) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, ByteOrder, u)
	if err == nil {
		n = 1
	}
	return
}

func (u Uint16) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint16(u))
}

func (u *Uint16) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, (*uint16)(u))
}

func (u Uint16) MarshalText() ([]byte, error) {
	return []byte(strconv.FormatUint(uint64(u), 10)), nil
}

func (u *Uint16) UnmarshalText(text []byte) error {
	v, err := strconv.ParseUint(string(text), 10, 16)
	if err != nil {
		return err
	}
	*u = Uint16(v)
	return nil
}

func (u Uint16) String() string {
	return strconv.FormatUint(uint64(u), 10)
}

type Uint32 uint32

func (Uint32) ObjectType() string {
	return "uint32"
}

func (u Uint32) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, ByteOrder, u)
	if err == nil {
		n = 1
	}
	return
}

func (u *Uint32) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, ByteOrder, u)
	if err == nil {
		n = 1
	}
	return
}

func (u Uint32) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint32(u))
}

func (u *Uint32) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, (*uint32)(u))
}

func (u Uint32) MarshalText() ([]byte, error) {
	return []byte(strconv.FormatUint(uint64(u), 10)), nil
}

func (u *Uint32) UnmarshalText(text []byte) error {
	v, err := strconv.ParseUint(string(text), 10, 32)
	if err != nil {
		return err
	}
	*u = Uint32(v)
	return nil
}

func (u Uint32) String() string {
	return strconv.FormatUint(uint64(u), 10)
}

type Uint64 uint64

func (Uint64) ObjectType() string {
	return "uint64"
}

func (u Uint64) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, ByteOrder, u)
	if err == nil {
		n = 1
	}
	return
}

func (u *Uint64) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, ByteOrder, u)
	if err == nil {
		n = 1
	}
	return
}

func (u Uint64) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint64(u))
}

func (u *Uint64) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, (*uint64)(u))
}

func (u Uint64) MarshalText() ([]byte, error) {
	return []byte(strconv.FormatUint(uint64(u), 10)), nil
}

func (u *Uint64) UnmarshalText(text []byte) error {
	v, err := strconv.ParseUint(string(text), 10, 64)
	if err != nil {
		return err
	}
	*u = Uint64(v)
	return nil
}

func (u Uint64) String() string {
	return strconv.FormatUint(uint64(u), 10)
}

type Int8 int8

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
