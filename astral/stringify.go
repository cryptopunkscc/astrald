package astral

import (
	"encoding"
	"fmt"
)

// Stringify renders any value as a string, trying Stringer, then TextMarshaler, then string, then fmt %v.
func Stringify(v any) string {
	if v == nil {
		return "nil"
	}

	if s, ok := v.(fmt.Stringer); ok {
		return s.String()
	}

	if r, ok := v.(encoding.TextMarshaler); ok {
		text, _ := r.MarshalText()
		return string(text)
	}

	if s, ok := v.(string); ok {
		return s
	}

	return fmt.Sprintf("%v", v)
}
