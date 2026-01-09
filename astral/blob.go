package astral

import (
	"encoding/base64"
	"encoding/json"
	"io"
)

// Blob is a raw bytes buffer with no type. Used to hold any arbitrary binary data as an Object.
type Blob []byte

var _ Object = (*Blob)(nil)

func (Blob) ObjectType() string {
	return "blob"
}

// binary

func (b Blob) WriteTo(w io.Writer) (_ int64, err error) {
	var m int
	m, err = w.Write(b)
	return int64(m), err
}

func (b *Blob) ReadFrom(r io.Reader) (n int64, err error) {
	var buf []byte
	buf, err = io.ReadAll(r)
	n = int64(len(buf))
	if err == nil {
		*b = buf
	}
	return
}

// json

func (b Blob) MarshalJSON() ([]byte, error) {
	return json.Marshal([]byte(b))
}

func (b *Blob) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, (*[]byte)(b))
}

// text

func (b Blob) MarshalText() (text []byte, err error) {
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(b)))
	base64.StdEncoding.Encode(buf, b)
	return buf, nil
}

func (b *Blob) UnmarshalText(text []byte) error {
	l := base64.StdEncoding.DecodedLen(len(text))
	*b = make(Blob, l)

	n, err := base64.StdEncoding.Decode(*b, text)
	if err != nil {
		return err
	}

	// trim the slice to the actual decoded length (in case of padding)
	*b = (*b)[:n]

	return nil
}

// ...

func (b Blob) String() string {
	return base64.StdEncoding.EncodeToString([]byte(b))
}

func init() {
	var b Blob
	_ = DefaultBlueprints.Add(&b)
}
