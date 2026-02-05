package astral

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"strconv"
)

type Float64 float64

var _ TextObject = (*Float64)(nil)

func NewFloat64(f float64) *Float64 {
	return (*Float64)(&f)
}

// astral:blueprint-ignore
func (Float64) ObjectType() string { return "float64" }

// binary

func (f Float64) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, ByteOrder, float64(f))
	if err == nil {
		n = 4
	}
	return
}

func (f *Float64) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, ByteOrder, f)
	if err == nil {
		n = 4
	}
	return
}

// json

func (f Float64) MarshalJSON() ([]byte, error) {
	return json.Marshal(float64(f))
}

func (f *Float64) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, (*float64)(f))
}

// text

func (f Float64) MarshalText() (text []byte, err error) {
	return []byte(strconv.FormatFloat(float64(f), 'f', -1, 64)), nil
}

func (f *Float64) UnmarshalText(text []byte) error {
	f64, err := strconv.ParseFloat(string(text), 64)
	if err != nil {
		return err
	}
	*f = Float64(f64)
	return nil
}

// ...

func init() {
	var f Float64
	Add(&f)
}
