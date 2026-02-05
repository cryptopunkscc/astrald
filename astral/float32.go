package astral

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"strconv"
)

type Float32 float32

var _ TextObject = (*Float32)(nil)

func NewFloat32(f float32) *Float32 {
	return (*Float32)(&f)
}

// astral:blueprint-ignore
func (Float32) ObjectType() string { return "float32" }

// binary

func (f Float32) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, ByteOrder, float32(f))
	if err == nil {
		n = 4
	}
	return
}

func (f *Float32) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, ByteOrder, f)
	if err == nil {
		n = 4
	}
	return
}

// json

func (f Float32) MarshalJSON() ([]byte, error) {
	return json.Marshal(float32(f))
}

func (f *Float32) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, (*float32)(f))
}

// text

func (f Float32) MarshalText() (text []byte, err error) {
	return []byte(strconv.FormatFloat(float64(f), 'f', -1, 32)), nil
}

func (f *Float32) UnmarshalText(text []byte) error {
	f64, err := strconv.ParseFloat(string(text), 32)
	if err != nil {
		return err
	}
	*f = Float32(f64)
	return nil
}

// ...

func init() {
	var f Float32
	Add(&f)
}
