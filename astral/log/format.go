package log

import (
	"fmt"
	"reflect"

	"github.com/cryptopunkscc/astrald/astral"
)

func Format(fmt string, v ...interface{}) (list []astral.Object) {
	var lastIndex int
	for i := 0; i < len(fmt) && len(v) > 0; i++ {
		if fmt[i] == '%' && i+1 < len(fmt) && fmt[i+1] == 'v' {
			// found a %v

			var s = astral.String32(fmt[lastIndex:i])
			if len(s) > 0 {
				list = append(list, &s)
			}

			var arg any
			arg, v = v[0], v[1:]

			list = append(list, toObject(arg))

			lastIndex = i + 2
			i++ // skip the next 'v'
		}
	}

	// flush the rest of the format string
	if lastIndex < len(fmt) {
		var s = astral.String32(fmt[lastIndex:])
		list = append(list, &s)
	}

	// add the rest of the objects
	for _, arg := range v {
		list = append(list, toObject(arg))
	}

	return
}

func toObject(v any) astral.Object {
	if o := astral.Adapt(v); o != nil {
		return o
	}

	if s, ok := v.(fmt.Stringer); ok {
		return astral.NewString32(s.String())
	}

	return astral.NewString32(reflect.TypeOf(v).String())
}
