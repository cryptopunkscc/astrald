package astral

import (
	"encoding/binary"
	"io"
	"strings"
)

type Bool bool

// astral

func (Bool) ObjectType() string {
	return "bool"
}

func (b Bool) WriteTo(w io.Writer) (n int64, err error) {
	err = binary.Write(w, encoding, b)
	if err == nil {
		n = 1
	}
	return
}

func (b *Bool) ReadFrom(r io.Reader) (n int64, err error) {
	err = binary.Read(r, encoding, b)
	if err == nil {
		n = 1
	}
	return
}

// text

func (b Bool) MarshalText() (text []byte, err error) {
	return []byte(b.String()), nil
}

func (b *Bool) UnmarshalText(text []byte) error {
	switch strings.ToLower(string(text)) {
	case "true", "yes", "t", "y", "1":
		*b = true
	case "false", "no", "f", "n", "0":
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
	_ = DefaultBlueprints.Add(&b)
}
