package astral

import (
	"fmt"
	"reflect"
	"strings"
)

func Stringify(v any) string {
	if v == nil {
		return "nil"
	}
	if s, ok := v.(fmt.Stringer); ok {
		return s.String()
	}
	return strings.TrimPrefix(reflect.TypeOf(v).String(), "*")
}
