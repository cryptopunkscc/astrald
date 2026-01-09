package astral

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
)

// Bytes8 is a byte buffer of 8-bit length
type Bytes8 []byte

// astral:blueprint-ignore
func (Bytes8) ObjectType() string {
	return "bytes8"
}

func (b Bytes8) WriteTo(w io.Writer) (n int64, err error) {
	var l = Uint8(len(b))
	if l > (1<<8)-1 {
		return 0, errors.New("data too large")
	}

	n, err = l.WriteTo(w)
	if err != nil {
		return
	}

	m, err := w.Write(b)
	n += int64(m)

	return
}

func (b *Bytes8) ReadFrom(r io.Reader) (n int64, err error) {
	var l Uint8
	n, err = l.ReadFrom(r)
	if err != nil {
		return
	}

	var buf = make([]byte, l)
	m, err := io.ReadFull(r, buf)
	n += int64(m)

	*b = Bytes8(buf[:m])

	return
}

func (b Bytes8) MarshalJSON() ([]byte, error) {
	return json.Marshal([]byte(b))
}

func (b *Bytes8) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, (*[]byte)(b))
}

func (b Bytes8) MarshalText() (text []byte, err error) {
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(b)))
	base64.StdEncoding.Encode(buf, b)
	return buf, nil
}

func (b *Bytes8) UnmarshalText(text []byte) error {
	l := base64.StdEncoding.DecodedLen(len(text))
	*b = make(Bytes8, l)

	n, err := base64.StdEncoding.Decode(*b, text)
	if err != nil {
		return err
	}

	// trim the slice to the actual decoded length (in case of padding)
	*b = (*b)[:n]

	return nil
}

// Bytes16 is a byte buffer of 16-bit length
type Bytes16 []byte

// astral:blueprint-ignore
func (Bytes16) ObjectType() string {
	return "bytes16"
}

func (b Bytes16) WriteTo(w io.Writer) (n int64, err error) {
	var l = Uint16(len(b))
	if l > (1<<16)-1 {
		return 0, errors.New("data too large")
	}

	n, err = l.WriteTo(w)
	if err != nil {
		return
	}

	m, err := w.Write(b)
	n += int64(m)

	return
}

func (b *Bytes16) ReadFrom(r io.Reader) (n int64, err error) {
	var l Uint16
	n, err = l.ReadFrom(r)
	if err != nil {
		return
	}

	var buf = make([]byte, l)
	m, err := io.ReadFull(r, buf)
	n += int64(m)

	*b = Bytes16(buf[:m])

	return
}

func (b Bytes16) MarshalJSON() ([]byte, error) {
	return json.Marshal([]byte(b))
}

func (b *Bytes16) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, (*[]byte)(b))
}

func (b Bytes16) MarshalText() (text []byte, err error) {
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(b)))
	base64.StdEncoding.Encode(buf, b)
	return buf, nil
}

func (b *Bytes16) UnmarshalText(text []byte) error {
	l := base64.StdEncoding.DecodedLen(len(text))
	*b = make(Bytes16, l)

	n, err := base64.StdEncoding.Decode(*b, text)
	if err != nil {
		return err
	}

	// trim the slice to the actual decoded length (in case of padding)
	*b = (*b)[:n]

	return nil
}

// Bytes32 is a byte buffer of 32-bit length
type Bytes32 []byte

// astral:blueprint-ignore
func (Bytes32) ObjectType() string {
	return "bytes32"
}

func (b Bytes32) WriteTo(w io.Writer) (n int64, err error) {
	var l = Uint32(len(b))
	if l > (1<<32)-1 {
		return 0, errors.New("data too large")
	}

	n, err = l.WriteTo(w)
	if err != nil {
		return
	}

	m, err := w.Write(b)
	n += int64(m)

	return
}

func (b *Bytes32) ReadFrom(r io.Reader) (n int64, err error) {
	var l Uint32
	n, err = l.ReadFrom(r)
	if err != nil {
		return
	}

	var buf = make([]byte, l)
	m, err := io.ReadFull(r, buf)
	n += int64(m)

	*b = Bytes32(buf[:m])

	return
}

func (b Bytes32) MarshalJSON() ([]byte, error) {
	return json.Marshal([]byte(b))
}

func (b *Bytes32) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, (*[]byte)(b))
}

func (b Bytes32) MarshalText() (text []byte, err error) {
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(b)))
	base64.StdEncoding.Encode(buf, b)
	return buf, nil
}

func (b *Bytes32) UnmarshalText(text []byte) error {
	l := base64.StdEncoding.DecodedLen(len(text))
	*b = make(Bytes32, l)

	n, err := base64.StdEncoding.Decode(*b, text)
	if err != nil {
		return err
	}

	// trim the slice to the actual decoded length (in case of padding)
	*b = (*b)[:n]

	return nil
}

// Bytes64 is a byte buffer of 64-bit length
type Bytes64 []byte

// astral:blueprint-ignore
func (Bytes64) ObjectType() string {
	return "bytes64"
}

func (b Bytes64) WriteTo(w io.Writer) (n int64, err error) {
	var l = Uint64(len(b))

	n, err = l.WriteTo(w)
	if err != nil {
		return
	}

	m, err := w.Write(b)
	n += int64(m)

	return
}

func (b *Bytes64) ReadFrom(r io.Reader) (n int64, err error) {
	var l Uint64
	n, err = l.ReadFrom(r)
	if err != nil {
		return
	}

	var buf = make([]byte, l)
	m, err := io.ReadFull(r, buf)
	n += int64(m)

	*b = Bytes64(buf[:m])

	return
}

func (b Bytes64) MarshalJSON() ([]byte, error) {
	return json.Marshal([]byte(b))
}

func (b *Bytes64) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, (*[]byte)(b))
}

func (b Bytes64) MarshalText() (text []byte, err error) {
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(b)))
	base64.StdEncoding.Encode(buf, b)
	return buf, nil
}

func (b *Bytes64) UnmarshalText(text []byte) error {
	l := base64.StdEncoding.DecodedLen(len(text))
	*b = make(Bytes64, l)

	n, err := base64.StdEncoding.Decode(*b, text)
	if err != nil {
		return err
	}

	// trim the slice to the actual decoded length (in case of padding)
	*b = (*b)[:n]

	return nil
}

func init() {
	var (
		b8  Bytes8
		b16 Bytes16
		b32 Bytes32
		b64 Bytes64
	)

	_ = DefaultBlueprints.Add(&b8, &b16, &b32, &b64)
}
