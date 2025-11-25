package astral

import (
	"io"
	"reflect"
)

// stringValue wraps a reflected string and uses astral.String32 for encoding
type stringValue struct {
	reflect.Value
}

var _ Object = &stringValue{}

func (stringValue) ObjectType() string {
	return String32("").ObjectType()
}

func (s stringValue) WriteTo(w io.Writer) (n int64, err error) {
	return String32(s.String()).WriteTo(w)
}

func (s stringValue) ReadFrom(r io.Reader) (n int64, err error) {
	var str String32
	n, err = str.ReadFrom(r)
	s.SetString(str.String())
	return
}
