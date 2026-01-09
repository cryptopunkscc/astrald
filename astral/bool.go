package astral

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"strings"
)

type Bool bool

func (Bool) ObjectType() string {
	return "bool"
}

// astral

func (b Bool) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, ByteOrder, b)
	if err == nil {
		n = 1
	}
	return
}

func (b *Bool) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, ByteOrder, b)
	if err == nil {
		n = 1
	}
	return
}

// json

func (b Bool) MarshalJSON() ([]byte, error) {
	return json.Marshal(bool(b))
}

func (b *Bool) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, (*bool)(b))
}

// text

func (b Bool) MarshalText() (text []byte, err error) {
	return []byte(b.String()), nil
}

func (b *Bool) UnmarshalText(text []byte) error {
	switch strings.ToLower(string(text)) {
	case "true", "yes", "t", "y":
		*b = true
	case "false", "no", "f", "n":
		*b = false
	default:
		return NewError("parse error")
	}
	return nil
}

// other

func (b Bool) String() string {
	if b {
		return "true"
	} else {
		return "false"
	}
}

func init() {
	var b Bool
	_ = Add(&b)
}
