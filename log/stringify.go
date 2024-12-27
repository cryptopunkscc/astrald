package log

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/sig"
	"reflect"
)

func Stringify(v any) string {
	if s, ok := v.(fmt.Stringer); ok {
		return s.String()
	}
	if reflect.TypeOf(v).Kind() == reflect.Pointer {
		return reflect.TypeOf(v).Elem().String()
	}
	return reflect.TypeOf(v).String()
}

func StringifySlice[T any](arr []T) (list []string) {
	list, _ = sig.MapSlice(arr, func(i T) (string, error) {
		return Stringify(i), nil
	})
	return
}
