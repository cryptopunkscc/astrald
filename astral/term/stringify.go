package term

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/sig"
	"reflect"
	"strconv"
)

func Stringify(v any) string {
	switch v := v.(type) {
	case string:
		return v
	case astral.Uint8:
		return strconv.FormatUint(uint64(v), 10)
	case astral.Uint16:
		return strconv.FormatUint(uint64(v), 10)
	case astral.Uint32:
		return strconv.FormatUint(uint64(v), 10)
	case astral.Uint64:
		return strconv.FormatUint(uint64(v), 10)
	case astral.Int8:
		return strconv.FormatInt(int64(v), 10)
	case astral.Int16:
		return strconv.FormatInt(int64(v), 10)
	case astral.Int32:
		return strconv.FormatInt(int64(v), 10)
	case astral.Int64:
		return strconv.FormatInt(int64(v), 10)
	case int:
		return strconv.FormatInt(int64(v), 10)
	}
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
