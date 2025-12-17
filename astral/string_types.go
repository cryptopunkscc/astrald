package astral

import (
	"encoding/json"
	"errors"
	"io"
)

// String represents a string of indefinite length
type String string

func (String) ObjectType() string {
	return "string"
}

func NewString(s string) *String {
	return (*String)(&s)
}

func (s String) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write([]byte(s))
	return int64(n), err
}

func (s *String) ReadFrom(r io.Reader) (n int64, err error) {
	var buf []byte
	buf, err = io.ReadAll(r)
	n = int64(len(buf))
	*s = String(buf)
	return

}

func (s String) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

func (s *String) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, (*string)(s))
}

func (s String) MarshalText() (text []byte, err error) {
	return []byte(s), nil
}

func (s *String) UnmarshalText(text []byte) error {
	*s = String(text)
	return nil
}

func (s String) String() string { return string(s) }

// String8 represents a string with an 8-bit length
type String8 string

func (String8) ObjectType() string {
	return "string8"
}

func (s String8) WriteTo(w io.Writer) (n int64, err error) {
	var l = Uint8(len(s))
	if l > (1<<8)-1 {
		return 0, errors.New("data too large")
	}

	n, err = l.WriteTo(w)
	if err != nil {
		return
	}

	m, err := w.Write([]byte(s))
	n += int64(m)

	return
}

func (s *String8) ReadFrom(r io.Reader) (n int64, err error) {
	var l Uint8
	n, err = l.ReadFrom(r)
	if err != nil {
		return
	}

	var buf = make([]byte, l)
	m, err := io.ReadFull(r, buf)
	n += int64(m)

	*s = String8(buf[:m])

	return

}

func (s String8) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

func (s *String8) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, (*string)(s))
}

func (s String8) MarshalText() (text []byte, err error) {
	return []byte(s), nil
}

func (s *String8) UnmarshalText(text []byte) error {
	*s = String8(text)
	return nil
}

func (s String8) String() string { return string(s) }

// String16 represents a string with a 16-bit length
type String16 string

func (String16) ObjectType() string {
	return "string16"
}

func (s String16) WriteTo(w io.Writer) (n int64, err error) {
	var l = Uint16(len(s))
	if l > (1<<16)-1 {
		return 0, errors.New("data too large")
	}

	n, err = l.WriteTo(w)
	if err != nil {
		return
	}

	m, err := w.Write([]byte(s))
	n += int64(m)

	return
}

func (s *String16) ReadFrom(r io.Reader) (n int64, err error) {
	var l Uint16
	n, err = l.ReadFrom(r)
	if err != nil {
		return
	}

	var buf = make([]byte, l)
	m, err := io.ReadFull(r, buf)
	n += int64(m)

	*s = String16(buf[:m])

	return
}

func (s String16) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

func (s *String16) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, (*string)(s))
}

func (s String16) MarshalText() (text []byte, err error) {
	return []byte(s), nil
}

func (s *String16) UnmarshalText(text []byte) error {
	*s = String16(text)
	return nil
}

func (s String16) String() string { return string(s) }

// String32 represents a string with a 32-bit length
type String32 string

func (String32) ObjectType() string {
	return "string32"
}

func (s String32) WriteTo(w io.Writer) (n int64, err error) {
	var l = Uint32(len(s))
	if l > (1<<32)-1 {
		return 0, errors.New("data too large")
	}

	n, err = l.WriteTo(w)
	if err != nil {
		return
	}

	m, err := w.Write([]byte(s))
	n += int64(m)

	return
}

func (s *String32) ReadFrom(r io.Reader) (n int64, err error) {
	var l Uint32
	n, err = l.ReadFrom(r)
	if err != nil {
		return
	}

	var buf = make([]byte, l)
	m, err := io.ReadFull(r, buf)
	n += int64(m)

	*s = String32(buf[:m])

	return
}

func (s String32) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

func (s *String32) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, (*string)(s))
}

func (s String32) MarshalText() (text []byte, err error) {
	return []byte(s), nil
}

func (s *String32) UnmarshalText(text []byte) error {
	*s = String32(text)
	return nil
}

func (s String32) String() string { return string(s) }

// String64 represents a string with a 64-bit length
type String64 string

func (String64) ObjectType() string { return "string64" }

func (s String64) WriteTo(w io.Writer) (n int64, err error) {
	var l = Uint64(len(s))

	n, err = l.WriteTo(w)
	if err != nil {
		return
	}

	m, err := w.Write([]byte(s))
	n += int64(m)

	return
}

func (s *String64) ReadFrom(r io.Reader) (n int64, err error) {
	var l Uint64
	n, err = l.ReadFrom(r)
	if err != nil {
		return
	}

	var buf = make([]byte, l)
	m, err := io.ReadFull(r, buf)
	n += int64(m)

	*s = String64(buf[:m])

	return
}

func (s String64) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

func (s *String64) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, (*string)(s))
}

func (s String64) MarshalText() (text []byte, err error) {
	return []byte(s), nil
}

func (s *String64) UnmarshalText(text []byte) error {
	*s = String64(text)
	return nil
}

func (s String64) String() string { return string(s) }

func init() {
	var (
		s   String
		s8  String8
		s16 String16
		s32 String32
		s64 String64
	)

	_ = DefaultBlueprints.Add(&s, &s8, &s16, &s32, &s64)
}
