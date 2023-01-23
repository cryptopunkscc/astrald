package log

import (
	"reflect"
)

type Formatter func(interface{}) string

var formatters map[string]Formatter

func Em(text string) string {
	return instance.Em(text)
}

func SetFormatter(t interface{}, formatter Formatter) {
	var typeName string
	if t == nil {
		typeName = "nil"
	} else {
		typeName = reflect.TypeOf(t).String()
	}

	formatters[typeName] = formatter
}

func formatArgs(args []interface{}) []interface{} {
	var out = make([]interface{}, 0)
	for _, a := range args {
		var argType = "nil"
		if reflect.TypeOf(a) != nil {
			argType = reflect.TypeOf(a).String()
		}

		if formatters[argType] != nil {
			f := formatters[argType](a)
			out = append(out, f)
			continue
		}

		if e, ok := a.(error); ok {
			out = append(out, red+e.Error()+reset)
			continue
		}

		out = append(out, a)
	}
	return out
}

func (l *Logger) Em(text string) string {
	return l.getEmColor() + text + reset
}

func init() {
	formatters = make(map[string]Formatter)

	SetFormatter(nil, func(i interface{}) string {
		return "nil"
	})
}
